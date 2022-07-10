// Package discserv provides a way to run an http server
// with logging and other necessary things
package discserv

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"crypto/ed25519"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/jdholdren/karma/internal/core"
	"go.uber.org/zap"
)

type Config struct {
	Port      int
	VerifyKey string
}

type Server struct {
	*http.Server

	cr  core.Core
	key ed25519.PublicKey
}

func New(l *zap.SugaredLogger, c Config, cr core.Core) (*Server, error) {
	r := mux.NewRouter()

	keyBytes, err := hex.DecodeString(c.VerifyKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding verify key: %s", err)
	}

	s := &Server{
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", c.Port),
			Handler:      r,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		cr:  cr,
		key: ed25519.PublicKey(keyBytes),
	}

	r.HandleFunc("/interactions", s.handleDiscordInteraction()).Methods(http.MethodPost)
	r.HandleFunc("/healthz", handleHealthCheck()).Methods(http.MethodGet)

	r.Use(loggingMiddleware(l))

	return s, nil
}

func loggingMiddleware(l *zap.SugaredLogger) mux.MiddlewareFunc {
	// God i hate the nesting
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Infow("request received", "uri", r.RequestURI, "method", r.Method)

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}

// What Discord sends us
type interaction struct {
	Type  uint            `json:"type"`
	Data  interactionData `json:"data"`
	Token string          `json:"token"`
}

type interactionData struct {
	Name     string              `json:"name"`
	Options  []interactionOption `json:"options"`
	Resolved resolvedData        `json:"resolved"`
}

type interactionOption struct {
	Name  string `json:"name"`
	Type  uint   `json:"type"`
	Value string `json:"value"`
}

type resolvedData struct {
	Users map[string]interactionUser `json:"users"`
}

type interactionUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (s *Server) handleDiscordInteraction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !discordgo.VerifyInteraction(r, s.key) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var i interaction
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			http.Error(w, fmt.Sprintf("error decoding: %s", err), http.StatusBadRequest)
			return
		}

		// Determine which handler to use
		if i.Type == 1 {
			s.handlePing(w)
			return
		}

		if i.Type == 2 && i.Data.Name == "gib" {
			s.handleGib(w, r, i.Data)
			return
		}
	}
}

func (s *Server) handlePing(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{ "type": 1 }`))
}

func (s *Server) handleGib(w http.ResponseWriter, r *http.Request, id interactionData) {
	givenID := id.Options[0].Value

	count, err := s.cr.AddKarma(r.Context(), givenID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error adding karma: %s", err), http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("You gave <@%s> karma. Their total is now %d", givenID, count.Count)

	w.Header().Add("Content-Type", "application/json")
	resp := fmt.Sprintf(`
	{
		"type": 4,
		"data": {
			"tts": false,
			"content": "%s",
			"embeds": []
		}
	}
	`, msg)
	w.Write([]byte(resp))
}

func handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
