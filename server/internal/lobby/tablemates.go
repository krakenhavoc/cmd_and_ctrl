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
//
// A deleted account is not a tablemate. Deleting an account scrubs the
// users row to a tombstone ("Deleted player") and leaves the seats
// pointing at it (accountdelete), so the join has to leave out
// `users.deleted_at IS NOT NULL` or the picker offers a person who can
// no longer be reached and the DM route fails on them with a 422.
//
// ADR 0213 §1 (slice 2) adds to each entry the number of tables the two
// people shared (`games_together`) and the name of the most recent one
// (`last_table`), and caps the list at the 50 most recent people
// (MaxTablemates). The cap is on what the list SHOWS: the DM route's
// gate (IsTablemate) reads the whole relation, so someone sat with long
// ago still counts as a tablemate for an invite.
//
// The same relation is the DM route's gate: POST /games/{id}/invites/dm
// lets an ordinary caller name only a person this query returns
// (IsTablemate), so the picker and the route cannot disagree about who
// counts.
//
// ADR 0213 slice 3 adds friends and personal blocks (internal/social).
// GET /me/tablemates leaves out anyone with a block between the two
// people, in either direction, and marks each entry as a friend or with
// a pending request (Config.playedWith). The DM route's gate becomes
// "a tablemate or a friend, and no block either way" (Config.mayReach),
// and so does the friend request's, which counts tablemates only.

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
	// GamesTogether is how many tables the two people have shared,
	// lobbies that never started included.
	GamesTogether int
	// LastTable is the name of the most recent shared table, the one
	// LastPlayedAt dates.
	LastTable string
}

// MaxTablemates caps GET /me/tablemates at the 50 most recent people
// (ADR 0213 §1).
const MaxTablemates = 50

// Tablemate is one entry of GET /me/tablemates. LastPlayedAt is Unix
// milliseconds, the unit the games table stores and the unit
// GET /me/games already answers in.
type Tablemate struct {
	UserID       uuid.UUID `json:"user_id"`
	DisplayName  string    `json:"display_name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	LastPlayedAt int64     `json:"last_played_at"`
	// GamesTogether is the count of tables the two of you have shared
	// (ADR 0213 §1); at least 1.
	GamesTogether int `json:"games_together"`
	// LastTable is the name of the most recent one, which is already in
	// both people's History.
	LastTable string `json:"last_table"`
	// Friend is whether the two of you are friends (ADR 0213 §1, slice
	// 3). Absent when you are not.
	Friend bool `json:"friend,omitempty"`
	// Request is a pending friend request between you: "incoming" (they
	// asked you, and you have not ignored it) or "outgoing" (you asked
	// them). Absent when there is none.
	Request string `json:"request,omitempty"`
}

// The values of Tablemate.Request.
const (
	RequestIncoming = "incoming"
	RequestOutgoing = "outgoing"
)

// Tablemates lists the people userID has shared a table with, most
// recent table first, at most MaxTablemates of them. The caller is
// never in the list, and a seat with no user — a guest, a bot, a
// Discord seat still waiting on a users row — is not a person and is
// skipped. So is a deleted account, which the SQL store leaves out (the
// memory store has no users table, so no account there is ever
// deleted).
func (l *Lobby) Tablemates(userID uuid.UUID) ([]Tablemate, error) {
	all, err := l.allTablemates(userID)
	if err != nil {
		return nil, err
	}
	if len(all) > MaxTablemates {
		all = all[:MaxTablemates]
	}
	return all, nil
}

// allTablemates is the whole relation, most recent first, for the
// callers that need membership rather than a page of it.
func (l *Lobby) allTablemates(userID uuid.UUID) ([]Tablemate, error) {
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
			UserID:        id,
			DisplayName:   r.DisplayName,
			AvatarURL:     r.AvatarURL,
			LastPlayedAt:  r.LastPlayedAt.UnixMilli(),
			GamesTogether: r.GamesTogether,
			LastTable:     r.LastTable,
		})
	}
	return out, nil
}

// IsTablemate reports whether target is one of userID's tablemates:
// the answer of Tablemates before its cap of MaxTablemates, so the
// picker and the invite-DM route share one definition of "someone you
// have sat at a table with" (a guest, a bot, yourself and a deleted
// account are not). It reads the whole list rather than asking the
// store a second question, which would be a second copy of the join;
// the list is the people one person has played with, which at this
// scale is small.
func (l *Lobby) IsTablemate(userID, target uuid.UUID) (bool, error) {
	if target == uuid.Nil {
		return false, nil
	}
	mates, err := l.allTablemates(userID)
	if err != nil {
		return false, err
	}
	for _, m := range mates {
		if m.UserID == target {
			return true, nil
		}
	}
	return false, nil
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
	mates, err := c.playedWith(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("listing a user's tablemates failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your tablemates; try again")
	}
	return writeJSON(w, http.StatusOK, tablematesResponse{Tablemates: mates})
}

// playedWith is the Played with list as GET /me/tablemates serves it
// (ADR 0213 §1, §3): the tablemates, minus anyone with a block between
// them and userID in either direction, each marked as a friend or with
// a pending request, capped at MaxTablemates. The block is applied
// before the cap, so a blocked person does not take one of the 50
// places.
func (c Config) playedWith(ctx context.Context, userID uuid.UUID) ([]Tablemate, error) {
	all, err := c.Lobby.allTablemates(userID)
	if err != nil {
		return nil, err
	}
	if c.Social != nil && len(all) > 0 {
		rel, err := c.Social.Relations(ctx, userID.String())
		if err != nil {
			return nil, err
		}
		kept := all[:0]
		for _, m := range all {
			id := m.UserID.String()
			if rel.Blocked[id] {
				continue
			}
			m.Friend = rel.Friends[id]
			switch {
			case rel.Incoming[id]:
				m.Request = RequestIncoming
			case rel.Outgoing[id]:
				m.Request = RequestOutgoing
			}
			kept = append(kept, m)
		}
		all = kept
	}
	if len(all) > MaxTablemates {
		all = all[:MaxTablemates]
	}
	return all, nil
}

// mayReach is the rule for naming another person in an invite or a
// friend request (ADR 0213 §1, §3): target is one of userID's tablemates
// (the whole relation, not the capped list) or friends, and neither has
// blocked the other. blockedByMe says the block that refused it was the
// caller's own, so the caller can be told; a block the other person made
// reads like any other refusal.
func (c Config) mayReach(ctx context.Context, userID, target uuid.UUID, friendsCount bool) (ok, blockedByMe bool, err error) {
	if target == uuid.Nil || target == userID {
		return false, false, nil
	}
	if c.Social != nil {
		mine, err := c.Social.IsBlocking(ctx, userID.String(), target.String())
		if err != nil {
			return false, false, err
		}
		if mine {
			return false, true, nil
		}
		theirs, err := c.Social.IsBlocking(ctx, target.String(), userID.String())
		if err != nil {
			return false, false, err
		}
		if theirs {
			return false, false, nil
		}
		if friendsCount {
			friends, err := c.Social.AreFriends(ctx, userID.String(), target.String())
			if err != nil {
				return false, false, err
			}
			if friends {
				return true, false, nil
			}
		}
	}
	mate, err := c.Lobby.IsTablemate(userID, target)
	return mate, false, err
}

// --- SQLStore ----------------------------------------------------

// Tablemates implements Store. One grouped self-join; see the file
// comment for why the recency anchor is games.created_at.
//
// The join onto `users` is an INNER join: seats.user_id is a foreign
// key to users(id) (migration 0003), so a non-NULL value always has a
// row, and an inner join keeps a corrupted id out of the picker
// rather than offering a nameless entry the DM route could not
// resolve. `u.deleted_at IS NULL` keeps a deleted account's tombstone
// out for the same reason: it has no Discord identity left to DM.
//
// `g.name` is a bare column in an aggregate query with one MAX(), so
// SQLite takes it from the row that holds the maximum: the name of the
// most recent shared table (SQLite's documented bare-column rule).
func (s *SQLStore) Tablemates(ctx context.Context, userID string) ([]TablemateRecord, error) {
	if userID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT s2.user_id, u.display_name, COALESCE(u.avatar_url, ''), MAX(g.created_at) AS last_played,
		        COUNT(DISTINCT g.id), g.name
		 FROM seats s1
		 JOIN seats s2 ON s2.game_id = s1.game_id
		 JOIN games g ON g.id = s1.game_id
		 JOIN users u ON u.id = s2.user_id
		 WHERE s1.user_id = ? AND s2.user_id IS NOT NULL AND s2.user_id <> ?
		   AND u.deleted_at IS NULL
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
		if err := rows.Scan(&rec.UserID, &rec.DisplayName, &rec.AvatarURL, &lastPlayed,
			&rec.GamesTogether, &rec.LastTable); err != nil {
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
	// COUNT(DISTINCT game) as the SQL store has it: one per shared
	// table, however many seats the person holds in it.
	counted := map[string]bool{}
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
				rec.LastTable = g.Name
			}
			if key := gid.String() + "|" + st.UserID; !counted[key] {
				counted[key] = true
				rec.GamesTogether++
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
