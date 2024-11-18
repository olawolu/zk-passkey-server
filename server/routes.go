package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	"github.com/olawolu/zk-pass/data"
)

func initRoutes(
	mux *mux.Router,
	config *Config,
	datastore *data.Datastore,
	sessionStore *SessionManager,
) {
	mux.Handle("/", http.NotFoundHandler())

	// register a new passkey
	registry := mux.PathPrefix("/register").Subrouter()
	registry.Handle("/initiate", beginRegistration(config, datastore, sessionStore))
	registry.HandleFunc("/finish", finishRegistration(config, datastore, sessionStore))

	// authenticate registered passkeys
	auth := mux.PathPrefix("auth").Subrouter()
	auth.HandleFunc("/initiate", beginLogin(config, datastore, sessionStore))
	auth.HandleFunc("/finish", finishLogin(config, datastore, sessionStore))
}

func beginRegistration(config *Config, datastore *data.Datastore, sessionStore *SessionManager) http.HandlerFunc {
	webAuthn, err := webauthn.New(config.webauthn)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user := datastore.GetUser() // Find or create the new user
		options, session, err := webAuthn.BeginRegistration(user)
		if err != nil {
			fmt.Println(err)
		}
		// store the sessionData values
		sessionStore.SaveSession(w, r, session)
		encode(w, http.StatusOK, options) // return the options generated
		// optionpublicKey contain our registration options
	}
}

func finishRegistration(config *Config, datastore *data.Datastore, sessionStore *SessionManager) http.HandlerFunc {
	webAuthn, err := webauthn.New(config.webauthn)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {

		user := datastore.GetUser() // Get the user

		// Get the session data stored from the function above
		session := sessionStore.GetSession(r, "key")

		credential, err := webAuthn.FinishRegistration(user, session, r)
		if err != nil {
			// Handle Error and return.
			return
		}

		// If creation was successful, store the credential object
		// Pseudocode to add the user credential.
		user.AddCredential(credential)
		datastore.SaveUser(user)

		encode(w, http.StatusOK, "Registration Success") // Handle next steps
	}
}

func beginLogin(config *Config, datastore *data.Datastore, sessionStore *SessionManager) http.HandlerFunc {
	webAuthn, err := webauthn.New(config.webauthn)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user := datastore.GetUser() // Find the user

		options, session, err := webAuthn.BeginLogin(user)
		if err != nil {
			// Handle Error and return.

			return
		}

		// store the session values
		sessionStore.SaveSession(w, r, session)

		encode(w, http.StatusOK, options) // return the options generated
		// optionpublicKey contain our registration options
	}
}

func finishLogin(config *Config, datastore *data.Datastore, sessionStore *SessionManager) http.HandlerFunc {
	webAuthn, err := webauthn.New(config.webauthn)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user := datastore.GetUser() // Get the user

		// Get the session data stored from the function above
		session := sessionStore.GetSession(r, "key")

		credential, err := webAuthn.FinishLogin(user, session, r)
		if err != nil {
			// Handle Error and return.
			return
		}

		// Handle credential.Authenticator.CloneWarning

		// If login was successful, update the credential object
		// Pseudocode to update the user credential.
		user.UpdateCredential(credential)
		datastore.SaveUser(user)

		encode(w, http.StatusOK, "Login Success")
	}
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
