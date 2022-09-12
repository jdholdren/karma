package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/jdholdren/karma/internal/core/models"
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

func (db DB) IncrementCount(ctx context.Context, userID string) error {
	q := `
	INSERT INTO karma_counts(user_id, count) VALUES (?, 1) ON CONFLICT(user_id) DO UPDATE SET count=count+1;
	`
	if _, err := db.db.ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("error incrementing karma_count: %s", err)
	}

	return nil
}

func (db DB) GetKarmaCount(ctx context.Context, userID string) (models.KarmaCount, error) {
	q := `
	SELECT * FROM karma_counts WHERE user_id = ?;
	`

	kc := models.KarmaCount{}
	if err := db.db.GetContext(ctx, &kc, q, userID); err != nil {
		return models.KarmaCount{}, fmt.Errorf("error incrementing karma_count: %s", err)
	}

	return kc, nil
}

func (db DB) TopCounts(ctx context.Context, n int64) ([]models.KarmaCount, error) {
	q := `
	SELECT * FROM karma_counts ORDER BY count DESC LIMIT ?;
	`

	kcs := []models.KarmaCount{}
	if err := db.db.SelectContext(ctx, &kcs, q, n); err != nil {
		return nil, fmt.Errorf("error querying top karma counts: %s", err)
	}

	return kcs, nil
}
