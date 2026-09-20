package lobby

// store_sql.go is the SQLite-backed Store: the games / seats / invites
// tables from migration 0002 (ADR 0051 decision 4). Hand-written SQL
// over database/sql, as internal/db asks.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// SQLStore is the Store production runs. Construct it with
// NewSQLStore over an opened, migrated *db.DB.
type SQLStore struct {
	db *db.DB
}

// NewSQLStore wraps an opened database. The caller owns d and closes
// it.
func NewSQLStore(d *db.DB) *SQLStore { return &SQLStore{db: d} }

// Durable is always true: rows outlive the process.
func (s *SQLStore) Durable() bool { return true }

// --- time and NULL helpers ---------------------------------------
//
// Every *_at column is Unix milliseconds (see migration 0002).

func toMillis(t time.Time) int64 { return t.UnixMilli() }

func fromMillis(ms int64) time.Time { return time.UnixMilli(ms).UTC() }

func nullMillis(t *time.Time) sql.NullInt64 {
	if t == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: t.UnixMilli(), Valid: true}
}

func timePtr(n sql.NullInt64) *time.Time {
	if !n.Valid {
		return nil
	}
	t := fromMillis(n.Int64)
	return &t
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullUUID(id uuid.UUID) sql.NullString {
	if id == uuid.Nil {
		return sql.NullString{}
	}
	return sql.NullString{String: id.String(), Valid: true}
}

func nullInt(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func intPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}

// execer is the part of *sql.DB and *sql.Tx the insert helpers use,
// so the importer can share them inside its transaction.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func insertGame(ctx context.Context, x execer, g GameRecord) error {
	_, err := x.ExecContext(ctx,
		`INSERT INTO games (id, name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat,
		                    host_player_id, host_discord_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		g.ID.String(), g.Name, nullString(g.CreatedBy), g.State, toMillis(g.CreatedAt),
		nullMillis(g.StartedAt), nullMillis(g.EndedAt), nullMillis(g.ArchivedAt), nullInt(g.WinnerSeat),
		nullUUID(g.HostPlayerID), nullString(g.HostDiscordID))
	return err
}

func insertSeats(ctx context.Context, x execer, gameID uuid.UUID, seats []SeatRecord) error {
	for _, st := range seats {
		if _, err := x.ExecContext(ctx,
			`INSERT INTO seats (game_id, seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			gameID.String(), st.Seat, st.PlayerID.String(), nullString(st.UserID), nullString(st.GuestName),
			nullString(st.BotTier), nullString(st.DeckID), nullString(st.DeckName), nullString(st.PendingDiscordID)); err != nil {
			return err
		}
	}
	return nil
}

func insertInvites(ctx context.Context, x execer, invites []InviteRecord) error {
	for _, inv := range invites {
		if _, err := x.ExecContext(ctx,
			`INSERT INTO invites (token_hash, game_id, kind, created_by, created_at, expires_at, revoked_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			inv.Hash[:], inv.GameID.String(), string(inv.Kind), nullString(inv.CreatedBy), toMillis(inv.CreatedAt),
			nullMillis(inv.ExpiresAt), nullMillis(inv.RevokedAt)); err != nil {
			return err
		}
	}
	return nil
}

// withTx runs fn in a transaction, committing only if it returns nil.
func (s *SQLStore) withTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLStore) CreateGame(ctx context.Context, g GameRecord, invites []InviteRecord) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		if err := insertGame(ctx, tx, g); err != nil {
			return fmt.Errorf("insert game: %w", err)
		}
		if err := insertInvites(ctx, tx, invites); err != nil {
			return fmt.Errorf("insert invites: %w", err)
		}
		return nil
	})
}

func (s *SQLStore) UpdateGame(ctx context.Context, g GameRecord) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE games SET name = ?, state = ?, started_at = ?, ended_at = ?, archived_at = ?, winner_seat = ?,
		                  host_player_id = ?, host_discord_id = ?
		 WHERE id = ?`,
		g.Name, g.State, nullMillis(g.StartedAt), nullMillis(g.EndedAt), nullMillis(g.ArchivedAt),
		nullInt(g.WinnerSeat), nullUUID(g.HostPlayerID), nullString(g.HostDiscordID), g.ID.String())
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrStoreNotFound
	}
	return nil
}

func (s *SQLStore) ReplaceSeats(ctx context.Context, gameID uuid.UUID, seats []SeatRecord) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM seats WHERE game_id = ?`, gameID.String()); err != nil {
			return err
		}
		return insertSeats(ctx, tx, gameID, seats)
	})
}

func (s *SQLStore) DeleteGame(ctx context.Context, id uuid.UUID) error {
	// Children first, explicitly: the ON DELETE CASCADE in the schema
	// is a backstop, not the mechanism.
	return s.withTx(ctx, func(tx *sql.Tx) error {
		for _, q := range []string{
			`DELETE FROM seats WHERE game_id = ?`,
			`DELETE FROM invites WHERE game_id = ?`,
			`DELETE FROM games WHERE id = ?`,
		} {
			if _, err := tx.ExecContext(ctx, q, id.String()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLStore) LoadGame(ctx context.Context, id uuid.UUID) (GameRecord, []SeatRecord, error) {
	var (
		g                              GameRecord
		createdBy                      sql.NullString
		createdAt                      int64
		startedAt, endedAt, archivedAt sql.NullInt64
		winner                         sql.NullInt64
		hostPlayer, hostDiscord        sql.NullString
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT name, created_by, state, created_at, started_at, ended_at, archived_at, winner_seat,
		        host_player_id, host_discord_id
		 FROM games WHERE id = ?`, id.String()).
		Scan(&g.Name, &createdBy, &g.State, &createdAt, &startedAt, &endedAt, &archivedAt, &winner,
			&hostPlayer, &hostDiscord)
	if errors.Is(err, sql.ErrNoRows) {
		return GameRecord{}, nil, ErrStoreNotFound
	}
	if err != nil {
		return GameRecord{}, nil, err
	}
	g.ID = id
	g.CreatedBy = createdBy.String
	g.CreatedAt = fromMillis(createdAt)
	g.StartedAt, g.EndedAt, g.ArchivedAt = timePtr(startedAt), timePtr(endedAt), timePtr(archivedAt)
	g.WinnerSeat = intPtr(winner)
	if hostPlayer.Valid {
		h, err := uuid.Parse(hostPlayer.String)
		if err != nil {
			return GameRecord{}, nil, fmt.Errorf("games.host_player_id %q: %w", hostPlayer.String, err)
		}
		g.HostPlayerID = h
	}
	g.HostDiscordID = hostDiscord.String

	rows, err := s.db.QueryContext(ctx,
		`SELECT seat, player_id, user_id, guest_name, bot_tier, deck_id, deck_name, pending_discord_id
		 FROM seats WHERE game_id = ? ORDER BY seat`, id.String())
	if err != nil {
		return GameRecord{}, nil, err
	}
	defer func() { _ = rows.Close() }()
	seats := []SeatRecord{}
	for rows.Next() {
		var (
			st                                                   SeatRecord
			playerID                                             string
			userID, guestName, botTier, deckID, deckName, discID sql.NullString
		)
		if err := rows.Scan(&st.Seat, &playerID, &userID, &guestName, &botTier, &deckID, &deckName, &discID); err != nil {
			return GameRecord{}, nil, err
		}
		pid, err := uuid.Parse(playerID)
		if err != nil {
			return GameRecord{}, nil, fmt.Errorf("seat %d player_id: %w", st.Seat, err)
		}
		st.PlayerID = pid
		st.UserID, st.GuestName, st.BotTier = userID.String, guestName.String, botTier.String
		st.DeckID, st.DeckName, st.PendingDiscordID = deckID.String, deckName.String, discID.String
		seats = append(seats, st)
	}
	if err := rows.Err(); err != nil {
		return GameRecord{}, nil, err
	}
	return g, seats, nil
}

func (s *SQLStore) GameIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM games`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []uuid.UUID
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("games.id %q: %w", raw, err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *SQLStore) Invite(ctx context.Context, hash InviteHash) (InviteRecord, error) {
	var (
		inv                  InviteRecord
		gameID, kind         string
		createdBy            sql.NullString
		createdAt            int64
		expiresAt, revokedAt sql.NullInt64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT game_id, kind, created_by, created_at, expires_at, revoked_at
		 FROM invites WHERE token_hash = ?`, hash[:]).
		Scan(&gameID, &kind, &createdBy, &createdAt, &expiresAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return InviteRecord{}, ErrStoreNotFound
	}
	if err != nil {
		return InviteRecord{}, err
	}
	id, err := uuid.Parse(gameID)
	if err != nil {
		return InviteRecord{}, fmt.Errorf("invites.game_id %q: %w", gameID, err)
	}
	inv.Hash = hash
	inv.GameID = id
	inv.Kind = InviteKind(kind)
	inv.CreatedBy = createdBy.String
	inv.CreatedAt = fromMillis(createdAt)
	inv.ExpiresAt, inv.RevokedAt = timePtr(expiresAt), timePtr(revokedAt)
	return inv, nil
}

func (s *SQLStore) RevokeInvite(ctx context.Context, hash InviteHash, at time.Time) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE invites SET revoked_at = ? WHERE token_hash = ?`, toMillis(at), hash[:])
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrStoreNotFound
	}
	return nil
}

func (s *SQLStore) RotateInvite(ctx context.Context, gameID uuid.UUID, kind InviteKind, newInvite InviteRecord, at time.Time) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		ok, err := gameExists(ctx, tx, gameID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrStoreNotFound
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE invites SET revoked_at = ? WHERE game_id = ? AND kind = ? AND revoked_at IS NULL`,
			toMillis(at), gameID.String(), string(kind)); err != nil {
			return fmt.Errorf("revoke live invites: %w", err)
		}
		if err := insertInvites(ctx, tx, []InviteRecord{newInvite}); err != nil {
			return fmt.Errorf("insert rotated invite: %w", err)
		}
		return nil
	})
}

// gameExists reports whether a games row exists, inside tx.
func gameExists(ctx context.Context, tx *sql.Tx, id uuid.UUID) (bool, error) {
	var one int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM games WHERE id = ?`, id.String()).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
