package lobby

// tablemates.go is ADR 0051 decision 8 — "tablemates are a query, not
// a table" (S34 sub-PR 6, tracking #607).
//
// The people you have shared a table with are a self-join of `seats`,
// nothing more: no friend requests, no acceptance, no `friendships`
// table. An explicit list, and the UI a social graph drags in, is
// deferred until the group is not eight people who already know each
// other; if it is ever wanted it is one table on top of this and
// changes nothing here.
//
// The query the ADR wrote, with the recency order it asked for and
// the display fields the picker needs:
//
//	SELECT DISTINCT s2.user_id
//	FROM seats s1 JOIN seats s2 ON s1.game_id = s2.game_id
//	WHERE s1.user_id = ? AND s2.user_id IS NOT NULL AND s2.user_id <> ?
//
// Recency is `games.created_at` of the most recent shared table.
// Not started_at or ended_at: those are NULL on a table that never
// started, and a lobby you both sat in yesterday is a better
// suggestion than a game you finished a year ago.

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
)

// TablemateRecord is one row of the tablemates query: a person the
// caller has shared at least one table with.
type TablemateRecord struct {
	// UserID is users(id) — our id, never a Discord snowflake.
	UserID string
	// DisplayName is the tablemate's current display name, so a
	// friend who renamed themselves on Discord reads correctly.
	DisplayName string
	// AvatarURL is the same-origin avatar path (users.avatar_url),
	// or "" for an account with no custom avatar.
	AvatarURL string
	// LastPlayedAt is when the most recent shared table was created.
	LastPlayedAt time.Time
}

// Tablemate is one entry of GET /me/tablemates. LastPlayedAt is Unix
// milliseconds, the unit the games table stores and the unit
// GET /me/games already answers in.
type Tablemate struct {
	UserID       uuid.UUID `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	LastPlayedAt int64     `json:"last_played_at"`
}

// Tablemates lists the people userID has shared a table with, most
// recent table first. The caller is never in the list, and a seat
// with no user — a guest, a bot, a Discord seat still waiting on a
// users row — is not a person and is skipped.
func (l *Lobby) Tablemates(userID uuid.UUID) ([]Tablemate, error) {
	out := []Tablemate{}
	if userID == uuid.Nil {
		return out, nil
	}
	ctx, cancel := storeCtx()
	recs, err := l.store.Tablemates(ctx, userID.String())
	cancel()
	if err != nil {
		return nil, err
	}
	for _, r := range recs {
		id, err := uuid.Parse(r.UserID)
		if err != nil {
			return nil, fmt.Errorf("seats.user_id %q: %w", r.UserID, err)
		}
		out = append(out, Tablemate{
			UserID:       id,
			DisplayName:  r.DisplayName,
			AvatarURL:    r.AvatarURL,
			LastPlayedAt: r.LastPlayedAt.UnixMilli(),
		})
	}
	return out, nil
}

// --- HTTP --------------------------------------------------------

// tablematesResponse is the body of GET /me/tablemates.
type tablematesResponse struct {
	Tablemates []Tablemate `json:"tablemates"`
}

// myTablemates is GET /me/tablemates: who to offer in the invite
// picker. Same caller rule as the rest of /me/*: a signed-in person,
// which on a deployment with no database is nobody.
func myTablemates(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	mates, err := c.Lobby.Tablemates(p.UserID)
	if err != nil {
		c.logger().Error("listing a user's tablemates failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your tablemates; try again")
	}
	return writeJSON(w, http.StatusOK, tablematesResponse{Tablemates: mates})
}

// --- SQLStore ----------------------------------------------------

// Tablemates implements Store. One grouped self-join; see the file
// comment for why the recency anchor is games.created_at.
//
// The join onto `users` is an INNER join: seats.user_id is a foreign
// key to users(id) (migration 0003), so a non-NULL value always has a
// row, and an inner join keeps a corrupted id out of the picker
// rather than offering a nameless entry the DM route could not
// resolve.
func (s *SQLStore) Tablemates(ctx context.Context, userID string) ([]TablemateRecord, error) {
	if userID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT s2.user_id, u.display_name, COALESCE(u.avatar_url, ''), MAX(g.created_at) AS last_played
		 FROM seats s1
		 JOIN seats s2 ON s2.game_id = s1.game_id
		 JOIN games g ON g.id = s1.game_id
		 JOIN users u ON u.id = s2.user_id
		 WHERE s1.user_id = ? AND s2.user_id IS NOT NULL AND s2.user_id <> ?
		 GROUP BY s2.user_id
		 ORDER BY last_played DESC, u.display_name, s2.user_id`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []TablemateRecord
	for rows.Next() {
		var (
			rec        TablemateRecord
			lastPlayed sql.NullInt64
		)
		if err := rows.Scan(&rec.UserID, &rec.DisplayName, &rec.AvatarURL, &lastPlayed); err != nil {
			return nil, err
		}
		if lastPlayed.Valid {
			rec.LastPlayedAt = fromMillis(lastPlayed.Int64)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// --- memoryStore -------------------------------------------------

// Tablemates implements Store. The memory store has no users table,
// so a tablemate's label is the name stored on the seat — the same
// compromise SeatsOfUser makes there.
func (s *memoryStore) Tablemates(_ context.Context, userID string) ([]TablemateRecord, error) {
	if userID == "" {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	byUser := map[string]TablemateRecord{}
	for gid, seats := range s.seats {
		g, ok := s.games[gid]
		if !ok {
			continue
		}
		mine := false
		for _, st := range seats {
			if st.UserID == userID {
				mine = true
				break
			}
		}
		if !mine {
			continue
		}
		for _, st := range seats {
			if st.UserID == "" || st.UserID == userID {
				continue
			}
			rec, seen := byUser[st.UserID]
			if !seen || g.CreatedAt.After(rec.LastPlayedAt) {
				rec.LastPlayedAt = g.CreatedAt
			}
			rec.UserID = st.UserID
			if st.GuestName != "" {
				rec.DisplayName = st.GuestName
			}
			byUser[st.UserID] = rec
		}
	}
	out := make([]TablemateRecord, 0, len(byUser))
	for _, rec := range byUser {
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if !a.LastPlayedAt.Equal(b.LastPlayedAt) {
			return a.LastPlayedAt.After(b.LastPlayedAt)
		}
		if a.DisplayName != b.DisplayName {
			return a.DisplayName < b.DisplayName
		}
		return a.UserID < b.UserID
	})
	return out, nil
}
