package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sethvargo/go-envconfig"

	"github.com/jdholdren/karma/internal/core"
	"github.com/jdholdren/karma/internal/core/db"
	"github.com/jdholdren/karma/internal/discord"
	"github.com/jdholdren/karma/internal/discserv"
	"github.com/jdholdren/karma/internal/logging"
)

//go:embed migrate/*
var f embed.FS

func main() {
	l := logging.NewLogger()
	defer func() {
		if err := l.Sync(); err != nil {
			log.Fatalf("error syncing logger: %s", err)
		}
	}()

	l.Debug("parsing config...")
	var cfg config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		l.Fatalf("error parsing config: %s", err)
	}
	l.Infow("parsed config", "config", cfg)

	// Connect to the database
	sqlDB, err := setupDB(cfg)
	if err != nil {
		l.Fatalf("error opening db: %s", err)
	}
	defer sqlDB.Close()
	d := db.New(sqlDB)

	cr := core.New(d)

	if !cfg.SkipRegister {
		dCli := discord.NewClient(
			discord.ClientConfig{
				AppID: cfg.DiscordAppID,
				Token: cfg.DiscordToken,
			},
			l.Named("discord_client"),
		)
		for _, guildID := range cfg.DiscordGuildIDs {
			if err := dCli.RegisterCommands(context.Background(), guildID); err != nil {
				l.Fatalf("error registering commands for guild '%s': %s", guildID, err)
			}
		}
	}

	s, err := discserv.New(
		l.Named("discserv"),
		discserv.Config{
			Port:      cfg.Port,
			VerifyKey: cfg.DiscordVerifyKey,
		},
		cr,
	)
	if err != nil {
		l.Fatalf("error creating discord server", "err", err)
	}

	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		err = s.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
	} else {
		err = s.ListenAndServe()
	}

	if err != nil {
		l.Errorw("error while serving", "err", err)
		return
	}
}

type config struct {
	// Server
	Port        int    `env:"PORT"`
	TLSCertFile string `env:"TLS_CERT_FILE"`
	TLSKeyFile  string `env:"TLS_KEY_FILE"`

	// Database
	DBPath string `env:"DB_PATH"`

	// Discord stuffs
	DiscordToken     string   `env:"DISCORD_TOKEN"`
	DiscordAppID     string   `env:"DISCORD_APP_ID"`
	DiscordGuildIDs  []string `env:"DISCORD_GUILD_IDS"`
	DiscordVerifyKey string   `env:"DISCORD_VERIFY_KEY"`
	// If we should not try to register commands with discord
	SkipRegister bool `env:"SKIP_REGISTER"`
}

// Connects to the db and migrates it
func setupDB(c config) (*sqlx.DB, error) {
	u, err := url.Parse(c.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error parsing db path: %s", err)
	}
	q := u.Query()
	q.Add("_journal", "WAL")
	u.RawQuery = q.Encode()

	db, err := sqlx.Open("sqlite3", u.String())
	if err != nil {
		return nil, fmt.Errorf("error opening db: %s", err)
	}

	// Perform migrations
	ups, err := f.ReadDir("migrate")
	if err != nil {
		return nil, fmt.Errorf("error reading migration dir: %s", err)
	}

	for _, up := range ups {
		if up.IsDir() {
			continue
		}

		if !strings.HasSuffix(up.Name(), "sql") {
			continue
		}

		upBytes, err := f.ReadFile(filepath.Join("migrate", up.Name()))
		if err != nil {
			return nil, fmt.Errorf("error reading up file: %s", err)
		}

		_, err = db.Exec(string(upBytes))
		if err != nil {
			return nil, fmt.Errorf("error executing up query for file %s: %s", up.Name(), err)
		}
	}

	return db, nil
}
