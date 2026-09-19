package lobby

// store_mygames.go is the person-shaped half of the lobby store (ADR
// 0051 sub-PR 4): linking the seats a Discord account claimed before
// it had a users row, and reading back every seat a user holds for
// "My games". Both implementations live here so the Store methods
// that exist for people are in one place.

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// --- SQLStore -----------------------------------------------------

// LinkPendingSeats implements Store. One UPDATE, in its own
// transaction: every seat waiting on the snowflake is linked, or none
// is.
func (s *SQLStore) LinkPendingSeats(ctx context.Context, discordID, userID string) (int, error) {
	if discordID == "" || userID == "" {
		return 0, nil
	}
	var n int64
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx,
			`UPDATE seats SET user_id = ?, pending_discord_id = NULL WHERE pending_discord_id = ?`,
			userID, discordID)
		if err != nil {
			return err
		}
		n, err = res.RowsAffected()
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("link pending seats: %w", err)
	}
	return int(n), nil
}

// SeatsOfUser implements Store. Two queries: the user's seats with
// their games, then every other seat at those tables. The label of
// another seat prefers its user's current display name, so a friend
// who renamed themselves on Discord reads correctly on old games too.
func (s *SQLStore) SeatsOfUser(ctx context.Context, userID string) ([]UserSeatRecord, error) {
	if userID == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.id, g.name, g.created_by, g.state, g.created_at, g.started_at, g.ended_at, g.archived_at, g.winner_seat, s.seat
		 FROM seats s JOIN games g ON g.id = s.game_id
		 WHERE s.user_id = ?
		 ORDER BY g.created_at DESC, g.id, s.seat`, userID)
	if err != nil {
		return nil, err
	}
	var out []UserSeatRecord
	index := map[uuid.UUID][]int{}
	for rows.Next() {
		var (
			rec                            UserSeatRecord
			id                             string
			createdBy                      sql.NullString
			createdAt                      int64
			startedAt, endedAt, archivedAt sql.NullInt64
			winner                         sql.NullInt64
		)
		if err := rows.Scan(&id, &rec.Game.Name, &createdBy, &rec.Game.State, &createdAt,
			&startedAt, &endedAt, &archivedAt, &winner, &rec.Seat); err != nil {
			_ = rows.Close()
			return nil, err
		}
		gid, err := uuid.Parse(id)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("games.id %q: %w", id, err)
		}
		rec.Game.ID = gid
		rec.Game.CreatedBy = createdBy.String
		rec.Game.CreatedAt = fromMillis(createdAt)
		rec.Game.StartedAt, rec.Game.EndedAt, rec.Game.ArchivedAt = timePtr(startedAt), timePtr(endedAt), timePtr(archivedAt)
		rec.Game.WinnerSeat = intPtr(winner)
		rec.Others = []OtherSeatRecord{}
		index[gid] = append(index[gid], len(out))
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()
	if len(out) == 0 {
		return out, nil
	}

	others, err := s.db.QueryContext(ctx,
		`SELECT o.game_id, o.seat, COALESCE(u.display_name, o.guest_name, ''), o.bot_tier IS NOT NULL
		 FROM seats o LEFT JOIN users u ON u.id = o.user_id
		 WHERE o.game_id IN (SELECT game_id FROM seats WHERE user_id = ?)
		   AND (o.user_id IS NULL OR o.user_id <> ?)
		 ORDER BY o.game_id, o.seat`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = others.Close() }()
	for others.Next() {
		var (
			gameID string
			o      OtherSeatRecord
			bot    int
		)
		if err := others.Scan(&gameID, &o.Seat, &o.Name, &bot); err != nil {
			return nil, err
		}
		o.Bot = bot != 0
		gid, err := uuid.Parse(gameID)
		if err != nil {
			return nil, fmt.Errorf("seats.game_id %q: %w", gameID, err)
		}
		for _, i := range index[gid] {
			out[i].Others = append(out[i].Others, o)
		}
	}
	return out, others.Err()
}

// --- memoryStore --------------------------------------------------

func (s *memoryStore) LinkPendingSeats(_ context.Context, discordID, userID string) (int, error) {
	if discordID == "" || userID == "" {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for gid, seats := range s.seats {
		for i := range seats {
			if seats[i].PendingDiscordID == discordID {
				seats[i].UserID = userID
				seats[i].PendingDiscordID = ""
				n++
			}
		}
		s.seats[gid] = seats
	}
	return n, nil
}

// SeatsOfUser implements Store. The memory store has no users table,
// so every label is the name stored on the seat.
func (s *memoryStore) SeatsOfUser(_ context.Context, userID string) ([]UserSeatRecord, error) {
	if userID == "" {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []UserSeatRecord
	for gid, seats := range s.seats {
		g, ok := s.games[gid]
		if !ok {
			continue
		}
		sorted := append([]SeatRecord(nil), seats...)
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].Seat < sorted[j].Seat })
		for _, mine := range sorted {
			if mine.UserID != userID {
				continue
			}
			rec := UserSeatRecord{Game: g, Seat: mine.Seat, Others: []OtherSeatRecord{}}
			for _, o := range sorted {
				if o.UserID == userID {
					continue
				}
				rec.Others = append(rec.Others, OtherSeatRecord{Seat: o.Seat, Name: o.GuestName, Bot: o.BotTier != ""})
			}
			out = append(out, rec)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if !a.Game.CreatedAt.Equal(b.Game.CreatedAt) {
			return a.Game.CreatedAt.After(b.Game.CreatedAt)
		}
		if a.Game.ID != b.Game.ID {
			return a.Game.ID.String() < b.Game.ID.String()
		}
		return a.Seat < b.Seat
	})
	return out, nil
}
