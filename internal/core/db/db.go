package db

import (
	"context"
	"fmt"

	"github.com/jdholdren/karma/internal/core/models"
	"github.com/jmoiron/sqlx"
)

// A DB struct holds the connection to sqlite and provides methods for interacting with
// persistent storage
type DB struct {
	db *sqlx.DB
}

// New creates an instance of our repository using the provided connection
func New(db *sqlx.DB) DB {
	return DB{
		db: db,
	}
}

func (db DB) IncrementCount(ctx context.Context, guildID, userID string) error {
	q := `
	INSERT INTO karma_counts(guild_id, user_id, count) VALUES (?, ?, 1) ON CONFLICT(guild_id, user_id) DO UPDATE SET count=count+1;
	`
	if _, err := db.db.ExecContext(ctx, q, guildID, userID); err != nil {
		return fmt.Errorf("error incrementing karma_count: %s", err)
	}

	return nil
}

func (db DB) GetKarmaCount(ctx context.Context, guildID, userID string) (models.KarmaCount, error) {
	q := `
	SELECT * FROM karma_counts WHERE guild_id = ? AND user_id = ? LIMIT 1;
	`

	kc := models.KarmaCount{}
	if err := db.db.GetContext(ctx, &kc, q, guildID, userID); err != nil {
		return models.KarmaCount{}, fmt.Errorf("error retrieving karma_count: %s", err)
	}

	return kc, nil
}

func (db DB) GetTopCountsForGuild(ctx context.Context, guildID string, top int) ([]models.KarmaCount, error) {
	q := `
	SELECT * FROM karma_counts WHERE guild_id = ? ORDER BY count DESC;
	`

	kcs := make([]models.KarmaCount, 0, top)
	if err := db.db.SelectContext(ctx, &kcs, q, guildID); err != nil {
		return nil, fmt.Errorf("error retrieving counts: %s", err)
	}

	return kcs, nil
}
