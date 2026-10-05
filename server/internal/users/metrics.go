package users

import (
	"context"
	"fmt"
	"time"
)

// The database half of the metrics users collector (ADR 0123 §3,
// metrics.UserSource): cmdctrl_users and cmdctrl_users_played. Read
// once a minute at most (metrics.UsersCacheTTL).

// CountUsers is the number of users rows: registered accounts.
func (s *SQLStore) CountUsers(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("users: count: %w", err)
	}
	return n, nil
}

// LastPlayed maps each account seated at a started table whose end is
// at or after since to the latest such end. A table's end is its
// ended_at, or its archived_at for one an operator closed by archiving
// it while it ran. A table with neither is not here: it is either live,
// which the lobby reports (metrics.LiveUserSource), or a row whose game
// did not come back after a restart, which nobody is playing.
//
// It reads seats.user_id and games' lifecycle columns, never
// users.last_seen_at, which is written only at sign-in.
func (s *SQLStore) LastPlayed(ctx context.Context, since time.Time) (map[string]time.Time, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.user_id, MAX(COALESCE(g.ended_at, g.archived_at))
		   FROM seats s JOIN games g ON g.id = s.game_id
		  WHERE s.user_id IS NOT NULL
		    AND g.started_at IS NOT NULL
		    AND COALESCE(g.ended_at, g.archived_at) >= ?
		  GROUP BY s.user_id`,
		since.UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("users: last played: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string]time.Time{}
	for rows.Next() {
		var (
			id  string
			end int64
		)
		if err := rows.Scan(&id, &end); err != nil {
			return nil, fmt.Errorf("users: last played: %w", err)
		}
		out[id] = time.UnixMilli(end)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("users: last played: %w", err)
	}
	return out, nil
}
