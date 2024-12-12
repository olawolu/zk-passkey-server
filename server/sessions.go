package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

type SessionManager struct {
	store sessions.Store
}

func NewSessionManager(store sessions.Store) *SessionManager {
	return &SessionManager{
		store: store,
	}
}
func (sm *SessionManager) GetSession(r *http.Request, key string) (*webauthn.SessionData, error) {
	session, err := sm.store.Get(r, key)
	if err != nil {
		err = fmt.Errorf("failed to get session: %w", err)
		return nil, err
	}

	return storeValueToSessionData(session.Values), nil
}

func (sm *SessionManager) SaveSession(w http.ResponseWriter, r *http.Request, sessionData *webauthn.SessionData, key string) error {
	session, err := sm.store.Get(r, key)
	if err != nil {
		err = fmt.Errorf("failed to get session: %w", err)
		return err
	}

	session.Values["challenge"] = sessionData.Challenge
	session.Values["rp"] = sessionData.RelyingPartyID
	session.Values["user_id"] = sessionData.UserID
	session.Values["allowed_creds"] = sessionData.AllowedCredentialIDs
	session.Values["expires"] = sessionData.Expires
	session.Values["user_ver"] = sessionData.UserVerification
	session.Values["extensions"] = sessionData.Extensions

	session.Options.MaxAge = 0
	if err = session.Save(r, w); err != nil {
		err = fmt.Errorf("failed to save session: %w", err)
		return err
	}
	return nil
}

// func sessionDataToStoreValues(sessionData *webauthn.SessionData) values map[interface{}]interface{}
func storeValueToSessionData(values map[interface{}]interface{}) *webauthn.SessionData {
	// userVerification, _ :=
	return &webauthn.SessionData{
		Challenge:            values["challenge"].(string),
		RelyingPartyID:       values["rp"].(string),
		UserID:               values["user_id"].([]byte),
		AllowedCredentialIDs: values["allowed_creds"].([][]byte),
		Expires:              values["expires"].(time.Time),
		UserVerification:     values["user_ver"].(protocol.UserVerificationRequirement),
		Extensions:           values["extensions"].(protocol.AuthenticationExtensions),
	}
}
