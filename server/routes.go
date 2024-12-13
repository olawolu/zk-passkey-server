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
	"github.com/olawolu/zk-pass/database"
	"github.com/olawolu/zk-pass/logger"
)

type RouteDoc struct {
    Path        string `json:"path"`
    Method      string `json:"method"`
    Description string `json:"description"`
}

var apiDocs = []RouteDoc{
    {
        Path:        "/register/initiate",
        Method:      "POST",
        Description: "Begin WebAuthn registration process. Returns credential creation options.",
    },
    {
        Path:        "/register/finish/{userId}",
        Method:      "POST", 
        Description: "Complete registration with attestation from authenticator.",
    },
    {
        Path:        "/login/initiate/{userId}",
        Method:      "POST",
        Description: "Begin WebAuthn authentication. Returns assertion options.",
    },
    {
        Path:        "/login/finish/{userId}",
        Method:      "POST",
        Description: "Complete authentication with assertion from authenticator.",
    },
}

type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func fmtResponse(code int, message string, data any) Response {
	return Response{
		Code:    code,
		Data:    data,
		Message: message,
	}
}

func initRoutes(
	mux *mux.Router,
	config *Config,
	datastore *database.DB,
	sessionStore *SessionManager,
	logger *logger.Logger,
) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<html>
			<head>
				<title>ZK-Passkey Server</title>
				<style>
					body { font-family: system-ui; max-width: 800px; margin: 0 auto; padding: 2rem; }
					.route { border: 1px solid #eee; padding: 1rem; margin: 1rem 0; border-radius: 4px; }
					.method { font-weight: bold; color: #0066cc; }
					.path { font-family: monospace; }
				</style>
			</head>
			<body>
				<h1>ZK-Passkey Server API</h1>
		`)
	
		for _, doc := range apiDocs {
			fmt.Fprintf(w, `
				<div class="route">
					<span class="method">%s</span>
					<span class="path">%s</span>
					<p>%s</p>
				</div>
			`, doc.Method, doc.Path, doc.Description)
		}
	
		fmt.Fprint(w, `
			</body>
		</html>
		`)
	})

	// register a new passkey
	registry := mux.PathPrefix("/register").Subrouter()
	registry.Handle("/initiate", beginRegistration(config, datastore, sessionStore, logger))
	registry.HandleFunc("/finish/{userId}", finishRegistration(config, datastore, sessionStore, logger))

	// authenticate registered passkeys
	auth := mux.PathPrefix("/login").Subrouter()
	auth.HandleFunc("/initiate/{userId}", beginLogin(config, datastore, sessionStore, logger))
	auth.HandleFunc("/finish/{userId}", finishLogin(config, datastore, sessionStore, logger))
}

func beginRegistration(
	config *Config,
	datastore *database.DB,
	sessionStore *SessionManager,
	log *logger.Logger,
) http.HandlerFunc {
	type registrationOptions struct {
		username string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		regOpts, err := decodeRequestBody[registrationOptions](r)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		user, err := datastore.RegisterNewUser(regOpts.username)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		creationOptions, session, err := webAuthn.BeginRegistration(user)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		sessionStore.SaveSession(w, r, session, user.ID.String())
		encodeJsonValue(w, http.StatusOK, creationOptions) // return the options generated
	}
}

func finishRegistration(
	config *Config,
	datastore *database.DB,
	sessionStore *SessionManager,
	log *logger.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userId := params["userId"]

		user, err := datastore.GetUser(userId) // Get the user
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		// Get the session data stored from the function above
		session, err := sessionStore.GetSession(r, userId)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		protocol.ParseCredentialCreationResponseBody(r.Body)
		credential, err := webAuthn.FinishRegistration(user, *session, r)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		err = datastore.AddCredential(credential, user.ID, r.UserAgent())
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		encodeJsonValue(w, http.StatusOK, "Registration Success") // Handle next steps
	}
}

func beginLogin(
	config *Config,
	datastore *database.DB,
	sessionStore *SessionManager,
	log *logger.Logger,
) http.HandlerFunc {
	type loginOptions struct {
		hashedTxIntent string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		userId := params["userId"]

		// loginOpts, err := decodeRequestBody[loginOptions](r)
		// if err != nil {
		// 	log.Logger.ErrorContext(r.Context(),err.Error())
		// }

		user, err := datastore.GetUser(userId) // Find the user
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		// challengeModifier := func(opt *protocol.PublicKeyCredentialRequestOptions) {
		// 	// opt.Challenge = loginOpts.hashedTxIntent{

		// }
		// webAuthn.G
		options, session, err := webAuthn.BeginLogin(user)
		if err != nil {
			// Handle Error and return.
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		// store the session values
		sessionStore.SaveSession(w, r, session, fmt.Sprintf("%s-%s", user.ID, user.PasskeyUserID))

		encodeJsonValue(w, http.StatusOK, options) // return the options generated
		// optionpublicKey contain our registration options
	}
}

func finishLogin(
	config *Config,
	datastore *database.DB,
	sessionStore *SessionManager,
	log *logger.Logger,
) http.HandlerFunc {
	type loginOptions struct {
		userId  string
		passkey string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		loginOpts, err := decodeRequestBody[loginOptions](r)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		user, err := datastore.GetUser(loginOpts.userId)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		session, err := sessionStore.GetSession(r, loginOpts.passkey)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		webAuthn, err := webauthn.New(config.webauthn)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}
		credential, err := webAuthn.FinishLogin(user, *session, r)
		if err != nil {
			log.Logger.ErrorContext(r.Context(), err.Error())
			response := fmtResponse(http.StatusInternalServerError, err.Error(), nil)
			encodeJsonValue[Response](w, http.StatusInternalServerError, response)
			return
		}

		// Handle credential.Authenticator.CloneWarning

		// If login was successful, update the credential object
		// Pseudocode to update the user credential.
		user.UpdateCredential(credential)
		encodeJsonValue(w, http.StatusOK, "Login Success")
	}
}

func encodeJsonValue[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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
