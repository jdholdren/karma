package core

import (
	"context"
	"fmt"

	"github.com/jdholdren/karma/internal/core/db"
	"github.com/jdholdren/karma/internal/core/models"
)

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
