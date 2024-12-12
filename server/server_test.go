package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/sessions"
	"github.com/olawolu/zk-pass/database"
	"github.com/olawolu/zk-pass/logger"
	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		port         string
		rpDisplay    string
		rpID         string
		rpOrigins    []string
		expectedConf *Config
	}{
		{
			name:      "valid config",
			host:      "localhost",
			port:      "8080",
			rpDisplay: "Test RP",
			rpID:      "localhost",
			rpOrigins: []string{"http://localhost:8080"},
			expectedConf: &Config{
				Host: "localhost",
				Port: "8080",
				webauthn: &webauthn.Config{
					RPDisplayName: "Test RP",
					RPID:          "localhost",
					RPOrigins:     []string{"http://localhost:8080"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ServerConfig(tt.host, tt.port, tt.rpDisplay, tt.rpID, tt.rpOrigins)
			assert.Equal(t, tt.expectedConf.Host, got.Host)
			assert.Equal(t, tt.expectedConf.Port, got.Port)
			assert.Equal(t, tt.expectedConf.webauthn.RPDisplayName, got.webauthn.RPDisplayName)
			assert.Equal(t, tt.expectedConf.webauthn.RPID, got.webauthn.RPID)
			assert.Equal(t, tt.expectedConf.webauthn.RPOrigins, got.webauthn.RPOrigins)
		})
	}
}

func TestNewServer(t *testing.T) {
	// Setup test dependencies
	// testLogger := logger.NewLogger()
	testDB := &database.DB{}
	testConfig, testLogger, testDB := createTestServer()
	testSessionStore := createTestSessionStore()
	// testConfig := &Config{
	// 	Host: "localhost",
	// 	Port: "8080",
	// 	webauthn: &webauthn.Config{
	// 		RPDisplayName: "Test RP",
	// 		RPID:          "localhost",
	// 		RPOrigins:     []string{"http://localhost:8080"},
	// 	},
	// }

	tests := []struct {
		name          string
		config        *Config
		logger        *logger.Logger
		db            *database.DB
		sessionStore  *SessionManager
		expectedPaths []string
	}{
		{
			name:         "creates server with routes",
			config:       testConfig,
			logger:       testLogger,
			db:           testDB,
			sessionStore: testSessionStore,
			expectedPaths: []string{
				"/register/initiate",
				"/register/finish/{userId}",
				"/login/initiate/{userId}",
				"/login/finish/{userId}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewServer(tt.config, tt.logger, tt.db, tt.sessionStore)

			// Test that handler is created
			assert.NotNil(t, handler)

			// Test routes are initialized
			for _, path := range tt.expectedPaths {
				req := httptest.NewRequest("POST", path, nil)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				fmt.Println(req.RequestURI, w.Code)
				// We expect 400 or similar error but not 404
				assert.NotEqual(t, http.StatusNotFound, w.Code)
			}
		})
	}
}

// Test helpers
func createTestServer() (*Config, *logger.Logger, *database.DB) {
	config := ServerConfig(
		"localhost",
		"8080",
		"Test RP",
		"localhost",
		[]string{"http://localhost:8080"},
	)
	logger := logger.NewLogger()
	db := &database.DB{}
	return config, logger, db
}

func createTestSessionStore() *SessionManager {
	store := sessions.NewFilesystemStore("")
	return NewSessionManager(store)
}
