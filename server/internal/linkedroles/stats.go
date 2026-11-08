package linkedroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Stats is one account's numbers (ADR 0163 §2) and the name its
// connection shows.
type Stats struct {
	// Name is users.display_name, the connection's platform_username.
	Name string
	// GamesPlayed counts finished games the account held a seat in:
	// state ended with an outcome of win or draw. An abandoned table
	// (outcome NULL) does not count, and a practice table has no row.
	GamesPlayed int
	// GamesWon counts those whose outcome is win and whose winner_seat is
	// the account's seat.
	GamesWon int
	// Joined is users.created_at.
	Joined time.Time
}

func (st Stats) connection() discord.RoleConnection {
	return discord.RoleConnection{
		PlatformName:     discord.RoleConnectionPlatform,
		PlatformUsername: st.Name,
		GamesPlayed:      st.GamesPlayed,
		GamesWon:         st.GamesWon,
		Joined:           st.Joined,
	}
}

// StatsReader computes an account's numbers. SQLStats is the real one.
type StatsReader interface {
	// Stats is ErrNotFound for no account, or a deleted one.
	Stats(ctx context.Context, userID uuid.UUID) (Stats, error)
}

// SQLStats reads Stats from the user database.
type SQLStats struct{ db *sql.DB }

// NewSQLStats wraps an opened, migrated database.
func NewSQLStats(d *db.DB) *SQLStats { return &SQLStats{db: d.DB} }

// statsQuery counts with the seat's own row: a bot or agent seat holds no
// account, so it counts for nobody, and a game is counted once however
// its seats are stored.
const statsQuery = `
SELECT u.display_name, u.created_at, u.deleted_at,
  (SELECT COUNT(DISTINCT g.id) FROM seats s JOIN games g ON g.id = s.game_id
    WHERE s.user_id = u.id AND s.bot_tier IS NULL AND s.agent_client IS NULL
      AND g.state = 'ended' AND g.outcome IN ('win', 'draw')),
  (SELECT COUNT(DISTINCT g.id) FROM seats s JOIN games g ON g.id = s.game_id
    WHERE s.user_id = u.id AND s.bot_tier IS NULL AND s.agent_client IS NULL
      AND g.state = 'ended' AND g.outcome = 'win' AND g.winner_seat = s.seat)
FROM users u WHERE u.id = ?`

// Stats implements StatsReader.
func (q *SQLStats) Stats(ctx context.Context, userID uuid.UUID) (Stats, error) {
	var (
		st        Stats
		createdAt int64
		deletedAt sql.NullInt64
	)
	err := q.db.QueryRowContext(ctx, statsQuery, userID.String()).
		Scan(&st.Name, &createdAt, &deletedAt, &st.GamesPlayed, &st.GamesWon)
	if errors.Is(err, sql.ErrNoRows) {
		return Stats{}, ErrNotFound
	}
	if err != nil {
		return Stats{}, fmt.Errorf("linkedroles: stats for %s: %w", userID, err)
	}
	if deletedAt.Valid {
		return Stats{}, ErrNotFound
	}
	st.Joined = time.UnixMilli(createdAt).UTC()
	return st, nil
}
