package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olawolu/zk-pass/data"
	"github.com/olawolu/zk-pass/logger"
)

func NewServer(
	config *Config,
	logger *logger.Logger,
	datastore *data.Datastore,
	sessionStore *SessionManager,
) http.Handler {
	mux := mux.NewRouter()
	initRoutes(mux, config, datastore, sessionStore)

	var handler http.Handler = mux
	// add some middleware
	return handler
}

// func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	s.router.ServeHTTP(w, r)
// }
