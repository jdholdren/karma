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

func New(c Config, cr core.Core) (*Server, error) {
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

	return s, nil
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
	username := id.Resolved.Users[id.Options[0].Value].Username

	count, err := s.cr.AddKarma(r.Context(), id.Options[0].Value)
	if err != nil {
		http.Error(w, fmt.Sprintf("error adding karma: %s", err), http.StatusInternalServerError)
		return
	}

	msg := fmt.Sprintf("You gave %s karma. Their total is now %d", username, count.Count)

	w.Header().Add("Content-Type", "application/json")
	resp := fmt.Sprintf(`
	{
		"type": 4,
		"data": {
			"tts": false,
			"content": "%s",
			"embeds": [],
			"allowed_mentions": { "parse": [] }
		}
	}
	`, msg)
	w.Write([]byte(resp))
}