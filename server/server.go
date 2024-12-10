package server

import (
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	data "github.com/olawolu/zk-pass/database"
	"github.com/olawolu/zk-pass/logger"
)

type Config struct {
	Host     string
	Port     string
	webauthn *webauthn.Config
}

// ServerConfig creates a new server configuration with the provided parameters.
//
// Parameters:
// - host: The hostname for the server.
// - port: The port number for the server.
// - rpDisplayName: The display name for your site, used in WebAuthn requests.
// - rpId: The relying party identifier, generally the fully qualified domain name (FQDN) for your site.
// - rpOrigins: A list of origin URLs allowed for WebAuthn requests.
//
// Returns:
// - A pointer to a Config struct containing the server configuration.
func ServerConfig(
	host string,
	port string,
	rpDisplayName string,
	rpId string,
	rpOrigins []string,
) *Config {
	wconfig := &webauthn.Config{
		RPDisplayName: rpDisplayName,
		RPID:          rpId,
		RPOrigins:     rpOrigins,
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
