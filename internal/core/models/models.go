// Package models provides the structs exposed by the core package,
// but put in an independent package to break the dependency cycle
// between `core` and `db`
package models

// A KarmaCount is a counter for karma attached to a user
type KarmaCount struct {
	GuildID string `db:"guild_id"`
	UserID  string `db:"user_id"`
	Count   uint   `db:"count"`
}
