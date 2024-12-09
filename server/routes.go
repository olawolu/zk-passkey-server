package server

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	"github.com/olawolu/zk-pass/data"
	"github.com/olawolu/zk-pass/logger"
)

func initRoutes(
	mux *mux.Router,
	config *Config,
	datastore *data.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) {
	mux.Handle("/", http.NotFoundHandler())

	// register a new passkey
	registry := mux.PathPrefix("/register").Subrouter()
	registry.Handle("/initiate", beginRegistration(config, datastore, sessionStore, logger))
	registry.HandleFunc("/finish/{userId}", finishRegistration(config, datastore, sessionStore, logger))

	// authenticate registered passkeys
	auth := mux.PathPrefix("login").Subrouter()
	auth.HandleFunc("/initiate/{userId}", beginLogin(config, datastore, sessionStore, logger))
	auth.HandleFunc("/finish/{userId}", finishLogin(config, datastore, sessionStore, logger))
}

func beginRegistration(
	config *Config,
	datastore *data.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) http.HandlerFunc {
	type registrationOptions struct {
		username string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		regOpts, err := decodeRequestBody[registrationOptions](r)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			encodeJsonValue(w, http.StatusInternalServerError, err.Error())
			return
		}

		user, err := datastore.RegisterNewUser(regOpts.username)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			encodeJsonValue(w, http.StatusInternalServerError, err.Error())
			return
		}

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			return
		}

		creationOptions, session, err := webAuthn.BeginRegistration(user)
		if err != nil {
			logger.Error(r.Context(), err.Error())
		}
		sessionStore.SaveSession(w, r, session, user.UserId.String())
		encodeJsonValue(w, http.StatusOK, creationOptions) // return the options generated
	}
}

func finishRegistration(
	config *Config,
	datastore *data.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userId := params["userId"]

		user, _ := datastore.GetUser(userId) // Get the user

		// Get the session data stored from the function above
		session := sessionStore.GetSession(r, userId)

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			return
		}
		protocol.ParseCredentialCreationResponseBody(r.Body)
		credential, err := webAuthn.FinishRegistration(user, session, r)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			// Handle Error and return.
			return
		}

		err = datastore.AddCredential(credential, user.UserId, r.UserAgent())
		if err != nil {
			logger.Error(r.Context(), err.Error())
			// Handle Error and return.
			return
		}
		encodeJsonValue(w, http.StatusOK, "Registration Success") // Handle next steps
	}
}

func beginLogin(
	config *Config,
	datastore *data.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) http.HandlerFunc {
	type loginOptions struct {
		hashedTxIntent string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userId := params["userId"]

		// loginOpts, err := decodeRequestBody[loginOptions](r)
		// if err != nil {
		// 	logger.Error(r.Context(),err.Error())
		// }

		user, _ := datastore.GetUser(userId) // Find the user

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			return
		}

		// challengeModifier := func(opt *protocol.PublicKeyCredentialRequestOptions) {
		// 	// opt.Challenge = loginOpts.hashedTxIntent{

		// }
		// webAuthn.G
		options, session, err := webAuthn.BeginLogin(user)
		if err != nil {
			// Handle Error and return.
			logger.Error(r.Context(), err.Error())
			return
		}
		// store the session values
		sessionStore.SaveSession(w, r, session, fmt.Sprintf("%s-%s", user.UserId, user.PasskeyUserId))

		encodeJsonValue(w, http.StatusOK, options) // return the options generated
		// optionpublicKey contain our registration options
	}
}

func finishLogin(
	config *Config,
	datastore *data.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) http.HandlerFunc {
	type loginOptions struct {
		userId  string
		passkey string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		loginOpts, err := decodeRequestBody[loginOptions](r)
		if err != nil {
			logger.Error(r.Context(), err.Error())
		}
		user, _ := datastore.GetUser(loginOpts.userId)
		session := sessionStore.GetSession(r, loginOpts.passkey)

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			fmt.Println(err)
			return
		}
		credential, err := webAuthn.FinishLogin(user, session, r)
		if err != nil {
			logger.Error(r.Context(), err.Error())
			return
		}

		// Handle credential.Authenticator.CloneWarning

		// If login was successful, update the credential object
		// Pseudocode to update the user credential.
		user.UpdateCredential(credential)
		datastore.SaveUser(user)

		encodeJsonValue(w, http.StatusOK, "Login Success")
	}
}

func encodeJsonValue[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decodeRequestBody[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func createChallenge(hashedTx []byte, nonce int) (challenge protocol.URLEncodedBase64, err error) {
	challenge = make([]byte, protocol.ChallengeLength)
	copy(challenge, hashedTx)
	nonceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(nonceBytes, uint32(nonce))
	copy(challenge[len(hashedTx):], nonceBytes)
	if _, err = rand.Read(challenge); err != nil {
		return nil, err
	}

	return challenge, nil
}
func createSecureChallenge(hashedTx []byte, nonce int) (protocol.URLEncodedBase64, error) {
	challenge := make([]byte, protocol.ChallengeLength)
	copy(challenge, hashedTx)
	nonceBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(nonceBytes, uint32(nonce))
	copy(challenge[len(hashedTx):], nonceBytes)
	if _, err := rand.Read(challenge[len(hashedTx)+len(nonceBytes):]); err != nil {
		return nil, err
	}
	return challenge, nil
}
