CREATE TABLE IF NOT EXISTS `karma_counts` (
  guild_id TEXT NOT NULL,
	user_id TEXT NOT NULL,
	count INTEGER NOT NULL,
  PRIMARY KEY(guild_id, user_id)
);
