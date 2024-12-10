package server

import (
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	"github.com/olawolu/zk-pass/database"
	"github.com/olawolu/zk-pass/logger"
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
		Host:     host,
		Port:     port,
		webauthn: wconfig,
	}
}

func NewServer(
	config *Config,
	logger *logger.Logger,
	datastore *data.DB,
) http.Handler {
	mux := mux.NewRouter()
	// sessionStore := NewSessionManager()
	initRoutes(mux, config, datastore, nil, logger)

	var handler http.Handler = mux
	// add some middleware
	return handler
}

// func loggingMiddleware(next http.Handler, logger *logger.Logger) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		logger.Info(,"Request received")
// 		next.ServeHTTP(w, r)
// 	})
// }
