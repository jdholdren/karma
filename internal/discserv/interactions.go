// Package discserv provides a way to run an http server
// with logging and other necessary things
package discserv

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crypto/ed25519"
	"crypto/tls"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/jdholdren/karma/internal/core"
	"go.uber.org/zap"
)

type Config struct {
	Port      int
	VerifyKey string

	TLSCertFile string
	TLSKeyFile  string
}

type Server struct {
	l *zap.SugaredLogger
	*http.Server

	cr  core.Core
	key ed25519.PublicKey // The discord public key to verify requests from them
}

func New(l *zap.SugaredLogger, c Config, cr core.Core) (*Server, error) {
	r := mux.NewRouter()

	keyBytes, err := hex.DecodeString(c.VerifyKey)
	if err != nil {
		return nil, fmt.Errorf("error decoding verify key: %s", err)
	}

	s := &Server{
		l: l,
		Server: &http.Server{
			Addr:         fmt.Sprintf(":%d", c.Port),
			Handler:      r,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		cr:  cr,
		key: ed25519.PublicKey(keyBytes),
	}

	if c.TLSCertFile != "" && c.TLSKeyFile != "" { // TLS key/cert provided
		l.Debug("setting up tls")
		loadKeyPair := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
			if err != nil {
				return nil, fmt.Errorf("error loading keypair for cert: %s", err)
			}

			return &cert, nil
		}
		s.TLSConfig = &tls.Config{
			GetCertificate: loadKeyPair,
		}
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
			if r.RequestURI == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			l.Infow("request received", "uri", r.RequestURI, "method", r.Method)

			// Call the next handler, which can be another middleware in the chain, or the final handler.
			next.ServeHTTP(w, r)
		})
	}
}

// What Discord sends us
type interaction struct {
	Type    uint            `json:"type"`
	Data    interactionData `json:"data"`
	GuildID string          `json:"guild_id"`
	Token   string          `json:"token"`
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
		l := s.l.With("method", "handleDiscordInteraction")

		if !discordgo.VerifyInteraction(r, s.key) {
			l.Debug("verification failed")
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		var i interaction
		if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
			l.Errorw("error decoding", "err", err)
			http.Error(w, fmt.Sprintf("error decoding: %s", err), http.StatusBadRequest)
			return
		}

		l.Infow("interaction decoded", "interaction", i)

		// Determine which handler to use
		if i.Type == 1 {
			s.handlePing(w)
			return
		}

		if i.Type == 2 && i.Data.Name == "gib" {
			s.handleGib(w, r, i)
			return
		}

		if i.Type == 2 && i.Data.Name == "checkkarma" {
			s.handleCheckKarma(w, r, i)
			return
		}

		if i.Type == 2 && i.Data.Name == "topten" {
			s.handleLeaderboard(w, r, i)
			return
		}
	}
}

func (s *Server) handlePing(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{ "type": 1 }`))
}

const (
	// A body with mentions
	mentionBody = `
	{
		"type": 4,
		"data": {
			"tts": false,
			"content": "%s",
			"embeds": []
		}
	}
	`
	mentionlessBody = `
	{
		"type": 4,
		"data": {
			"tts": false,
			"content": "%s",
			"embeds": [],
			"allowed_mentions": {
				"parse": []
			}
		}
	}
	`
)

func writeMsgResponse(w http.ResponseWriter, message string, allowMentions bool) {
	w.Header().Add("Content-Type", "application/json")
	body := mentionBody
	if !allowMentions {
		body = mentionlessBody
	}
	resp := fmt.Sprintf(body, message)
	_, _ = w.Write([]byte(resp))
}

func (s *Server) handleGib(w http.ResponseWriter, r *http.Request, i interaction) {
	guildID := i.GuildID
	givenID := i.Data.Options[0].Value
	msg := i.Data.Options[1].Value

	count, err := s.cr.AddKarma(r.Context(), guildID, givenID)
	if err != nil {
		s.l.Errorw("error adding karma", "err", err)
		http.Error(w, fmt.Sprintf("error adding karma: %s", err), http.StatusInternalServerError)
		return
	}

	s.l.Infow("sucessfully added karma", "given_to", givenID)

	content := fmt.Sprintf("You gave <@%s> karma for '%s'. Their total is now %d", givenID, msg, count.Count)
	writeMsgResponse(w, content, true)
}

func (s *Server) handleCheckKarma(w http.ResponseWriter, r *http.Request, i interaction) {
	userID := i.Data.Options[0].Value
	username := i.Data.Resolved.Users[i.Data.Options[0].Value].Username

	count, err := s.cr.GetKarma(r.Context(), i.GuildID, userID)
	if err != nil {
		s.l.Errorw("error checking karma", "err", err)
		http.Error(w, fmt.Sprintf("error checking karma: %s", err), http.StatusInternalServerError)
		return
	}

	s.l.Infow("sucessfully checked karma", "username", username)

	content := fmt.Sprintf("Checked %s's karma. Their total is %d", username, count.Count)
	writeMsgResponse(w, content, false)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request, i interaction) {
	counts, err := s.cr.GetTopCounts(r.Context(), i.GuildID, 10)
	if err != nil {
		s.l.Errorw("error checking leaderboard", "err", err)
		http.Error(w, fmt.Sprintf("error checking leaderboard: %s", err), http.StatusInternalServerError)
		return
	}

	b := &strings.Builder{}
	for j, count := range counts {
		b.WriteString(fmt.Sprintf("%d. <@%s>: %d karma \\n", j+1, count.UserID, count.Count))
	}

	writeMsgResponse(w, b.String(), false)
}

func handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
