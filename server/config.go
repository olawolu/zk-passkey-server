package server

import (
	"github.com/go-webauthn/webauthn/webauthn"
)

type Config struct {
	Host     string
	Port     string
	webauthn *webauthn.Config
}

func ServerConfig(host, port string) *Config {
	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",                               // Display Name for your site
		RPID:          "go-webauthn.local",                         // Generally the FQDN for your site
		RPOrigins:     []string{"https://login.go-webauthn.local"}, // The origin URLs allowed for WebAuthn requests
	}
	return &Config{
		Host:     "",
		Port:     "",
		webauthn: wconfig,
	}
}
