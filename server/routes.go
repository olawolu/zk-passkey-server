package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	user := s.datastore.GetUser() // Find or create the new user
	options, session, err := s.webAuthn.BeginRegistration(user)
	if err != nil {
		fmt.Println(err)
	}
	// store the sessionData values
	s.session.SaveSession(w, r, session)
	JSONResponse(w, options, http.StatusOK) // return the options generated
	// options.publicKey contain our registration options
}

func (s *Server) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	user := s.datastore.GetUser() // Get the user

	// Get the session data stored from the function above
	session := s.session.GetSession(r, "key")

	credential, err := s.webAuthn.FinishRegistration(user, session, r)
	if err != nil {
		// Handle Error and return.
		return
	}

	// If creation was successful, store the credential object
	// Pseudocode to add the user credential.
	user.AddCredential(credential)
	s.datastore.SaveUser(user)

	JSONResponse(w, "Registration Success", http.StatusOK) // Handle next steps
}

func (s *Server) BeginLogin(w http.ResponseWriter, r *http.Request) {
	user := s.datastore.GetUser() // Find the user

	options, session, err := s.webAuthn.BeginLogin(user)
	if err != nil {
		// Handle Error and return.

		return
	}

	// store the session values
	s.session.SaveSession(w, r, session)

	JSONResponse(w, options, http.StatusOK) // return the options generated
	// options.publicKey contain our registration options
}

func (s *Server) FinishLogin(w http.ResponseWriter, r *http.Request) {
	user := s.datastore.GetUser() // Get the user

	// Get the session data stored from the function above
	session := s.session.GetSession(r, "key")

	credential, err := s.webAuthn.FinishLogin(user, session, r)
	if err != nil {
		// Handle Error and return.

		return
	}

	// Handle credential.Authenticator.CloneWarning

	// If login was successful, update the credential object
	// Pseudocode to update the user credential.
	user.UpdateCredential(credential)
	s.datastore.SaveUser(user)

	JSONResponse(w, "Login Success", http.StatusOK)
}

func JSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code
	w.WriteHeader(statusCode)

	// Encode the data as JSON and write it to the response
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
