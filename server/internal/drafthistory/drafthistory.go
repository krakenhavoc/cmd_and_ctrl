// Package drafthistory keeps a signed-in person's finished drafts so
// they can look back at one (ADR 0222 §9, #5440): the set, the date,
// the cards they picked, the deck they built and how their matches
// went. It is migration 0038.
//
// A draft itself is never stored: it lives in the lobby's memory and is
// gone when it closes. A guest's draft keeps nothing at all; only a
// seat with an account writes a row here.
//
// This is personal data (ADR 0149). Account deletion calls
// ForgetAccount inside its own transaction.
package drafthistory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// ErrNotFound: the person has no record of that draft.
var ErrNotFound = errors.New("drafthistory: not found")

// Pick is one card a person picked.
type Pick struct {
	Name        string   `json:"name"`
	ScryfallID  string   `json:"scryfall_id"`
	Set         string   `json:"set"`
	CollectorNo string   `json:"collector_number"`
	Rarity      string   `json:"rarity"`
	Colors      []string `json:"colors"`
	ManaCost    string   `json:"mana_cost"`
	ManaValue   int      `json:"mana_value"`
	TypeLine    string   `json:"type_line"`
	Basic       bool     `json:"basic,omitempty"`
	// InDeck says the person's deck played the card; the rest was their
	// sideboard.
	InDeck bool `json:"in_deck"`
}

// The kinds of opponent a stored match names.
const (
	OpponentPerson = "person"
	OpponentGuest  = "guest"
	OpponentBot    = "bot"
	OpponentBye    = "bye"
)

// Match is one round a person played.
type Match struct {
	Round int `json:"round"`
	// OpponentKind is one of the Opponent constants.
	OpponentKind string `json:"opponent_kind"`
	// OpponentUser is the opponent's account when it was a signed-in
	// person, uuid.Nil otherwise. It is never sent to a client.
	OpponentUser uuid.UUID `json:"-"`
	// OpponentName is that account's display name now, filled on a
	// read, "" when there is none to show.
	OpponentName string `json:"opponent_name,omitempty"`
	// Result is "win", "loss" or "draw", "" for a bye.
	Result    string `json:"result"`
	GamesWon  int    `json:"games_won"`
	GamesLost int    `json:"games_lost"`
}

// Record is one person's look back at one draft.
type Record struct {
	DraftID   uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"-"`
	Name      string    `json:"name"`
	Set       string    `json:"set"`
	SetName   string    `json:"set_name"`
	Mode      string    `json:"mode"`
	Seats     int       `json:"seats"`
	BestOf    int       `json:"best_of"`
	Seat      int       `json:"seat"`
	DraftedAt time.Time `json:"drafted_at"`
	// Picks and Basics are absent from a list (Store.List).
	Picks   []Pick         `json:"picks,omitempty"`
	Basics  map[string]int `json:"basics,omitempty"`
	Matches []Match        `json:"matches"`
	// Wins, Losses and Draws are counted from Matches on a read.
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

// Store keeps draft records. Safe for concurrent use.
type Store interface {
	// Save writes one person's record of one draft, replacing an
	// earlier one: a draft is saved when a person leaves it and again
	// when it finishes. A record with no UserID is refused.
	Save(ctx context.Context, r Record) error
	// List is a person's drafts, newest first, without their cards.
	List(ctx context.Context, userID uuid.UUID) ([]Record, error)
	// Get is one draft of a person's, with its cards. ErrNotFound when
	// they have none by that id: a record is its owner's alone.
	Get(ctx context.Context, userID, draftID uuid.UUID) (Record, error)
}

// ListLimit is the most drafts List returns.
const ListLimit = 100

// SQLStore is migration 0038's tables.
type SQLStore struct{ db *sql.DB }

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore { return &SQLStore{db: d.DB} }

var _ Store = (*SQLStore)(nil)

// Save implements Store.
func (s *SQLStore) Save(ctx context.Context, r Record) error {
	if r.UserID == uuid.Nil || r.DraftID == uuid.Nil {
		return fmt.Errorf("drafthistory: a record needs a draft and an account")
	}
	if r.Picks == nil {
		r.Picks = []Pick{}
	}
	picks, err := json.Marshal(r.Picks)
	if err != nil {
		return err
	}
	if r.Basics == nil {
		r.Basics = map[string]int{}
	}
	basics, err := json.Marshal(r.Basics)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	did, uid := r.DraftID.String(), r.UserID.String()
	if _, err := tx.ExecContext(ctx, `DELETE FROM draft_history WHERE draft_id = ? AND user_id = ?`, did, uid); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO draft_history (draft_id, user_id, draft_name, set_code, set_name, mode, seats, best_of, seat, drafted_at, picks, basics)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		did, uid, r.Name, r.Set, r.SetName, r.Mode, r.Seats, r.BestOf, r.Seat, r.DraftedAt.UnixMilli(), string(picks), string(basics)); err != nil {
		return err
	}
	for _, m := range r.Matches {
		var opp any
		if m.OpponentKind == OpponentPerson && m.OpponentUser != uuid.Nil {
			opp = m.OpponentUser.String()
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO draft_history_matches (draft_id, user_id, round, opponent_kind, opponent_user_id, result, games_won, games_lost)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			did, uid, m.Round, m.OpponentKind, opp, m.Result, m.GamesWon, m.GamesLost); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// matches reads the rounds of one record, naming each opponent by the
// display name their account has now.
func (s *SQLStore) matches(ctx context.Context, draftID, userID string) ([]Match, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT m.round, m.opponent_kind, COALESCE(u.display_name, ''), m.result, m.games_won, m.games_lost
		   FROM draft_history_matches m
		   LEFT JOIN users u ON u.id = m.opponent_user_id
		  WHERE m.draft_id = ? AND m.user_id = ?
		  ORDER BY m.round`, draftID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Match{}
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.Round, &m.OpponentKind, &m.OpponentName, &m.Result, &m.GamesWon, &m.GamesLost); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func tally(r *Record) {
	r.Wins, r.Losses, r.Draws = 0, 0, 0
	for _, m := range r.Matches {
		switch m.Result {
		case "win":
			r.Wins++
		case "loss":
			r.Losses++
		case "draw":
			r.Draws++
		}
	}
}

const recordColumns = `draft_id, draft_name, set_code, set_name, mode, seats, best_of, seat, drafted_at`

func scanRecord(scan func(...any) error, r *Record, extra ...any) error {
	var id string
	var at int64
	dest := append([]any{&id, &r.Name, &r.Set, &r.SetName, &r.Mode, &r.Seats, &r.BestOf, &r.Seat, &at}, extra...)
	if err := scan(dest...); err != nil {
		return err
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	r.DraftID = parsed
	r.DraftedAt = time.UnixMilli(at).UTC()
	return nil
}

// List implements Store.
func (s *SQLStore) List(ctx context.Context, userID uuid.UUID) ([]Record, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+recordColumns+` FROM draft_history WHERE user_id = ? ORDER BY drafted_at DESC, draft_id LIMIT ?`,
		userID.String(), ListLimit)
	if err != nil {
		return nil, err
	}
	out := []Record{}
	for rows.Next() {
		r := Record{UserID: userID}
		if err := scanRecord(rows.Scan, &r); err != nil {
			_ = rows.Close()
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if out[i].Matches, err = s.matches(ctx, out[i].DraftID.String(), userID.String()); err != nil {
			return nil, err
		}
		tally(&out[i])
	}
	return out, nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, userID, draftID uuid.UUID) (Record, error) {
	r := Record{UserID: userID}
	var picks, basics string
	row := s.db.QueryRowContext(ctx,
		`SELECT `+recordColumns+`, picks, basics FROM draft_history WHERE draft_id = ? AND user_id = ?`,
		draftID.String(), userID.String())
	if err := scanRecord(row.Scan, &r, &picks, &basics); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, err
	}
	if err := json.Unmarshal([]byte(picks), &r.Picks); err != nil {
		return Record{}, fmt.Errorf("drafthistory: picks of %s: %w", draftID, err)
	}
	if err := json.Unmarshal([]byte(basics), &r.Basics); err != nil {
		return Record{}, fmt.Errorf("drafthistory: basics of %s: %w", draftID, err)
	}
	var err error
	if r.Matches, err = s.matches(ctx, draftID.String(), userID.String()); err != nil {
		return Record{}, err
	}
	tally(&r)
	return r, nil
}

// AccountCounts is what ForgetAccount removed.
type AccountCounts struct {
	// Drafts is the person's own records, Matches the rounds of them,
	// and Opponents the rounds of other people's records that named the
	// person and no longer do.
	Drafts, Matches, Opponents int
}

// ForgetAccount is account deletion's part (ADR 0149, ADR 0222 §9),
// inside the deletion's own transaction: the person's draft records
// and their rounds are deleted, and wherever another person's record
// names them as an opponent the account id is cleared. That round then
// reads as played against a person with no name.
func ForgetAccount(ctx context.Context, tx *sql.Tx, user string) (AccountCounts, error) {
	var c AccountCounts
	run := func(into *int, q string, args ...any) error {
		res, err := tx.ExecContext(ctx, q, args...)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		*into = int(n)
		return err
	}
	// The rounds first: a database opened without foreign keys does
	// not cascade them.
	if err := run(&c.Matches, `DELETE FROM draft_history_matches WHERE user_id = ?`, user); err != nil {
		return c, err
	}
	if err := run(&c.Drafts, `DELETE FROM draft_history WHERE user_id = ?`, user); err != nil {
		return c, err
	}
	if err := run(&c.Opponents, `UPDATE draft_history_matches SET opponent_user_id = NULL WHERE opponent_user_id = ?`, user); err != nil {
		return c, err
	}
	return c, nil
}
