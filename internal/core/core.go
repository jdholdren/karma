package core

import (
	"context"
	"fmt"

	"github.com/jdholdren/karma/internal/core/db"
	"github.com/jdholdren/karma/internal/core/models"
)

type ValidationErr struct {
	Field string
	Msg   string
}

func (err ValidationErr) Error() string {
	return fmt.Sprintf("error with '%s': %s", err.Field, err.Msg)
}

type Core struct {
	db db.DB
}

func New(db db.DB) Core {
	return Core{
		db: db,
	}
}

// AddKarma increments the karma for a user
func (c Core) AddKarma(ctx context.Context, userID string) (models.KarmaCount, error) {
	if err := c.db.IncrementCount(ctx, userID); err != nil {
		return models.KarmaCount{}, fmt.Errorf("error incrementing count: %s", err)
	}

	count, err := c.db.GetKarmaCount(ctx, userID)
	if err != nil {
		return models.KarmaCount{}, fmt.Errorf("error getting count: %s", err)
	}

	return count, nil
}

func (c Core) GetKarma(ctx context.Context, userID string) (models.KarmaCount, error) {
	count, err := c.db.GetKarmaCount(ctx, userID)
	if err != nil {
		return models.KarmaCount{}, fmt.Errorf("error getting count: %s", err)
	}

	return count, nil
}

func (c Core) TopCounts(ctx context.Context, n int64) ([]models.KarmaCount, error) {
	// Don't let them check more than 10
	if n > 10 || n < 1 {
		return nil, ValidationErr{
			Field: "num",
			Msg:   "must provide valid number between 1 and 10",
		}
	}

	counts, err := c.db.TopCounts(ctx, n)
	if err != nil {
		return nil, fmt.Errorf("error getting top db counts: %s", err)
	}

	return counts, nil
}
