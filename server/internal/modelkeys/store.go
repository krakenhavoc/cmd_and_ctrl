package modelkeys

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// record is one model_keys row (migration 0025). Sealed is the only
// form of the key that is ever stored.
type record struct {
	UserID       uuid.UUID
	Provider     string
	BaseURL      string
	Model        string
	Sealed       []byte
	Last4        string
	GamesPerDay  int
	CallsPerGame int
	Day          string
	DayGames     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastUsedAt   time.Time
	LastError    string
	LastErrorAt  time.Time
}

// sqlStore is the model_keys table.
type sqlStore struct {
	db *sql.DB
}

func newSQLStore(d *db.DB) *sqlStore { return &sqlStore{db: d.DB} }

func fromMs(v sql.NullInt64) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return time.UnixMilli(v.Int64).UTC()
}

func (s *sqlStore) get(ctx context.Context, user uuid.UUID) (record, bool, error) {
	var (
		r                   record
		day, lastErr        sql.NullString
		created, updated    int64
		lastUsed, lastErrAt sql.NullInt64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game, day, day_games,
		        created_at, updated_at, last_used_at, last_error, last_error_at
		 FROM model_keys WHERE user_id = ?`, user.String()).
		Scan(&r.Provider, &r.BaseURL, &r.Model, &r.Sealed, &r.Last4, &r.GamesPerDay, &r.CallsPerGame, &day, &r.DayGames,
			&created, &updated, &lastUsed, &lastErr, &lastErrAt)
	if errors.Is(err, sql.ErrNoRows) {
		return record{}, false, nil
	}
	if err != nil {
		// Never wrap the row: a scan error can quote a value.
		return record{}, false, errors.New("modelkeys: read the key row failed")
	}
	r.UserID = user
	r.Day, r.LastError = day.String, lastErr.String
	r.CreatedAt, r.UpdatedAt = time.UnixMilli(created).UTC(), time.UnixMilli(updated).UTC()
	r.LastUsedAt, r.LastErrorAt = fromMs(lastUsed), fromMs(lastErrAt)
	return r, true, nil
}

// put inserts or replaces the person's row. A replacement keeps the
// original created_at and resets the error, which belonged to the old
// configuration.
func (s *sqlStore) put(ctx context.Context, r record) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO model_keys (user_id, provider, base_url, model, sealed_key, last4, games_per_day, calls_per_game,
		                         day, day_games, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, 0, ?, ?)
		 ON CONFLICT (user_id) DO UPDATE SET
		   provider = excluded.provider, base_url = excluded.base_url, model = excluded.model,
		   sealed_key = excluded.sealed_key, last4 = excluded.last4,
		   games_per_day = excluded.games_per_day, calls_per_game = excluded.calls_per_game,
		   updated_at = excluded.updated_at, last_error = NULL, last_error_at = NULL`,
		r.UserID.String(), r.Provider, r.BaseURL, r.Model, r.Sealed, r.Last4, r.GamesPerDay, r.CallsPerGame,
		r.CreatedAt.UnixMilli(), r.UpdatedAt.UnixMilli())
	if err != nil {
		return fmt.Errorf("modelkeys: save the key row failed: %w", sanitizeDBErr(err))
	}
	return nil
}

func (s *sqlStore) remove(ctx context.Context, user uuid.UUID) (bool, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM model_keys WHERE user_id = ?`, user.String())
	if err != nil {
		return false, fmt.Errorf("modelkeys: delete the key row failed: %w", sanitizeDBErr(err))
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (s *sqlStore) touchUsed(ctx context.Context, user uuid.UUID, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE model_keys SET last_used_at = ? WHERE user_id = ?`, at.UnixMilli(), user.String())
	return sanitizeDBErr(err)
}

// setError records a failure. msg is always one of this package's fixed
// sentences.
func (s *sqlStore) setError(ctx context.Context, user uuid.UUID, msg string, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE model_keys SET last_error = ?, last_error_at = ? WHERE user_id = ?`,
		msg, at.UnixMilli(), user.String())
	return sanitizeDBErr(err)
}

func (s *sqlStore) clearError(ctx context.Context, user uuid.UUID, at time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE model_keys SET last_error = NULL, last_error_at = NULL, last_used_at = ? WHERE user_id = ?`,
		at.UnixMilli(), user.String())
	return sanitizeDBErr(err)
}

// claimGame counts one more game on today's UTC date, if the owner's
// games-a-day limit allows it. It reports whether it did. One statement,
// so two seats starting at once cannot both take the last game.
func (s *sqlStore) claimGame(ctx context.Context, user uuid.UUID, day string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE model_keys
		    SET day_games = CASE WHEN day = ? THEN day_games + 1 ELSE 1 END, day = ?
		  WHERE user_id = ? AND (day IS NOT ? OR day_games < games_per_day)`,
		day, day, user.String(), day)
	if err != nil {
		return false, sanitizeDBErr(err)
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

// gamesToday is how many games have used the key on day.
func (s *sqlStore) gamesToday(ctx context.Context, user uuid.UUID, day string) (used, limit int, ok bool, err error) {
	var d sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT day, day_games, games_per_day FROM model_keys WHERE user_id = ?`, user.String()).
		Scan(&d, &used, &limit)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, sanitizeDBErr(err)
	}
	if d.String != day {
		used = 0
	}
	return used, limit, true, nil
}

// sanitizeDBErr keeps a database error's kind but not its text: a
// constraint failure message can quote a value, and the sealed key is a
// value. The sealed key is not the key, but nothing about it belongs in
// a log either.
func sanitizeDBErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return errors.New("database error")
}
