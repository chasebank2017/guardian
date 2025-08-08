package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Agent represents the structure of the agents table.
type Agent struct {
	ID         int
	Hostname   string
	OsVersion  string
	Status     string
	LastSeenAt time.Time
	CreatedAt  time.Time
}

// CreateAgent inserts a new agent into the database and returns the new agent's ID.
func CreateAgent(ctx context.Context, pool *pgxpool.Pool, hostname, osVersion string) (int, error) {
	var agentID int
	query := `
		INSERT INTO agents (hostname, os_version, status, last_seen_at)
		VALUES ($1, $2, 'online', NOW())
		RETURNING id
	`
	err := pool.QueryRow(ctx, query, hostname, osVersion).Scan(&agentID)
	if err != nil {
		return 0, err
	}
	return agentID, nil
}
