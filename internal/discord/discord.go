package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client is the struct that provides interactivity with discord
type Client struct {
	appID      string
	token      string // The secret token
	httpClient *http.Client

	l *zap.SugaredLogger
}

type ClientConfig struct {
	AppID string
	Token string
}

// NewClient produces a new client with the given config
func NewClient(c ClientConfig, l *zap.SugaredLogger) *Client {
	return &Client{
		appID: c.AppID,
		token: c.Token,
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		l: l,
	}
}

func (c *Client) setupRequest(r *http.Request) {
	r.Header.Add("Authorization", fmt.Sprintf("Bot %s", c.token))
	r.Header.Add("Content-Type", "application/json")
}

// Represents the request to create a command
type command struct {
	Name        string          `json:"name"`
	Type        uint            `json:"type"`
	Description string          `json:"description"`
	Options     []commandOption `json:"options"`
}

type commandOption struct {
	Name        string `json:"name"`
	Type        uint   `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// RegisterCommands reaces out to discord to register all commands supported by the app
func (c *Client) RegisterCommands(ctx context.Context, guildID string) error {
	cmds := []command{
		{
			Name:        "gib",
			Type:        1, // CHAT_INPUT
			Description: "Give another user one karma",
			Options: []commandOption{
				{
					Name:        "user",
					Type:        6, // USER
					Description: "The user to give karma to",
					Required:    true,
				},
				{
					Name:        "message",
					Type:        3, // STRING
					Description: "Message to accompany the gifting of karma",
					Required:    true,
				},
			},
		},
		{
			Name:        "checkkarma",
			Type:        1, // CHAT_INPUT
			Description: "Check a user's karma",
			Options: []commandOption{
				{
					Name:        "user",
					Type:        6, // USER
					Description: "The user to check",
					Required:    true,
				},
			},
		},
	}

	// The give karma command
	for _, cmd := range cmds {
		if err := c.registerCommand(ctx, cmd, guildID); err != nil {
			return fmt.Errorf("error registering command %s: %s", cmd.Name, err)
		}
	}

	return nil
}

func (c *Client) registerCommand(ctx context.Context, cmd command, guildID string) error {
	byts, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("error marshalling gib command: %s", err)
	}

	u := fmt.Sprintf("https://discord.com/api/v10/applications/%s/guilds/%s/commands", c.appID, guildID)
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(byts))
	if err != nil {
		return fmt.Errorf("error creating request to create gib command: %s", err)
	}
	req = req.WithContext(ctx)
	c.setupRequest(req)

	c.l.Debugw("calling to register commands", "url", u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error doing request: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		er, err := readErr(res.Body)
		if err != nil {
			return fmt.Errorf("error reading error from body: %s", err)
		}

		c.l.Errorw("received error response from api", "err", er, "status_code", res.StatusCode)
		return er
	}

	c.l.Infow("sucessfully called to register guild command", "command_name", cmd.Name)

	return nil
}
