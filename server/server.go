package server

import (
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/olawolu/zk-pass/data"
)

type Server struct {
	webAuthn  *webauthn.WebAuthn
	datastore *data.Datastore
	session   *SessionManager
}

func NewServer() *Server {
	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",                               // Display Name for your site
		RPID:          "go-webauthn.local",                         // Generally the FQDN for your site
		RPOrigins:     []string{"https://login.go-webauthn.local"}, // The origin URLs allowed for WebAuthn requests
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &Server{
		webAuthn: webAuthn,
	}
}
