package gamestats

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

// ErrNotFound: the game has no record.
var ErrNotFound = errors.New("gamestats: not found")

// Store keeps game records. Safe for concurrent use.
type Store interface {
	// Put replaces the game's record, its seats and its cards, in one
	// transaction. The game must have a games row.
	Put(ctx context.Context, g Game) error
	// Get reads one game's record, seats in seat order and each seat's
	// cards by name. ErrNotFound if there is none.
	Get(ctx context.Context, id uuid.UUID) (Game, error)
	// ForUser summarises the finished games of the seats whose
	// seats.user_id is userID, and no one else's (ADR 0165 §7).
	ForUser(ctx context.Context, userID string) (UserStats, error)
	// Overview summarises every game for an operator (ADR 0165 §8).
	Overview(ctx context.Context) (Overview, error)
}

// NoStore is the Store for a deployment with no database: Put keeps
// nothing and Get finds nothing.
type NoStore struct{}

// Put implements Store; it does nothing.
func (NoStore) Put(context.Context, Game) error { return nil }

// Get implements Store; there is never a record.
func (NoStore) Get(context.Context, uuid.UUID) (Game, error) { return Game{}, ErrNotFound }

// ForUser implements Store; there is nothing to summarise.
func (NoStore) ForUser(context.Context, string) (UserStats, error) {
	return UserStats{ByDeck: []Group{}, ByCommander: []Group{}, Recent: []RecentGame{}, TopCards: []TopCard{}}, nil
}

var _ Store = NoStore{}

// SQLStore is migration 0023's game_stats, seat_stats and
// seat_card_stats. computed_at is Unix milliseconds.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore { return &SQLStore{db: d.DB} }

// Put implements Store.
func (s *SQLStore) Put(ctx context.Context, g Game) (err error) {
	if g.GameID == uuid.Nil {
		return errors.New("gamestats: game id is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("gamestats: begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	id := g.GameID.String()
	// The seats and cards go with the game row (ON DELETE CASCADE).
	if _, err = tx.ExecContext(ctx, `DELETE FROM game_stats WHERE game_id = ?`, id); err != nil {
		return fmt.Errorf("gamestats: clear: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO game_stats (game_id, stats_version, source, log_complete, turns, rounds, starting_seat, computed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, g.Version, string(g.Source), boolInt(g.LogComplete), g.Turns, g.Rounds, nullInt(g.StartingSeat),
		g.ComputedAt.Truncate(time.Millisecond).UnixMilli()); err != nil {
		return fmt.Errorf("gamestats: put game: %w", err)
	}
	for _, st := range g.Seats {
		commanders, merr := json.Marshal(nonNil(st.Commanders))
		if merr != nil {
			return fmt.Errorf("gamestats: commanders: %w", merr)
		}
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO seat_stats (game_id, seat, result, place, eliminated_turn, conceded, final_life,
			   life_gained, life_lost, damage_to_players, combat_damage_to_players, commander_damage_dealt,
			   commander_damage_taken, spells_cast, cards_drawn, attacks, tokens_created, mulligans, commanders)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, st.Seat, string(st.Result), nullInt(st.Place), nullInt(st.EliminatedTurn), boolInt(st.Conceded),
			st.FinalLife, st.LifeGained, st.LifeLost, st.DamageToPlayers, st.CombatDamageToPlayers,
			st.CommanderDamageDealt, st.CommanderDamageTaken, st.SpellsCast, st.CardsDrawn, st.Attacks,
			st.TokensCreated, st.Mulligans, string(commanders)); err != nil {
			return fmt.Errorf("gamestats: put seat %d: %w", st.Seat, err)
		}
		for _, c := range st.Cards {
			if _, err = tx.ExecContext(ctx,
				`INSERT INTO seat_card_stats (game_id, seat, name, scryfall_id, casts, first_cast_turn)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				id, st.Seat, c.Name, nullString(c.ScryfallID), c.Casts, c.FirstCastTurn); err != nil {
				return fmt.Errorf("gamestats: put seat %d card %q: %w", st.Seat, c.Name, err)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("gamestats: commit: %w", err)
	}
	return nil
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, id uuid.UUID) (Game, error) {
	g := Game{GameID: id}
	var (
		source   string
		complete int
		start    sql.NullInt64
		computed int64
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT stats_version, source, log_complete, turns, rounds, starting_seat, computed_at
		 FROM game_stats WHERE game_id = ?`, id.String()).
		Scan(&g.Version, &source, &complete, &g.Turns, &g.Rounds, &start, &computed)
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrNotFound
	}
	if err != nil {
		return Game{}, fmt.Errorf("gamestats: get game: %w", err)
	}
	g.Source, g.LogComplete, g.StartingSeat = Source(source), complete == 1, intPtr(start)
	g.ComputedAt = time.UnixMilli(computed).UTC()

	rows, err := s.db.QueryContext(ctx,
		`SELECT seat, result, place, eliminated_turn, conceded, final_life, life_gained, life_lost,
		   damage_to_players, combat_damage_to_players, commander_damage_dealt, commander_damage_taken,
		   spells_cast, cards_drawn, attacks, tokens_created, mulligans, commanders
		 FROM seat_stats WHERE game_id = ? ORDER BY seat`, id.String())
	if err != nil {
		return Game{}, fmt.Errorf("gamestats: get seats: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var (
			st              Seat
			result, cmdrs   string
			place, elimTurn sql.NullInt64
			conceded        int
		)
		if err := rows.Scan(&st.Seat, &result, &place, &elimTurn, &conceded, &st.FinalLife, &st.LifeGained,
			&st.LifeLost, &st.DamageToPlayers, &st.CombatDamageToPlayers, &st.CommanderDamageDealt,
			&st.CommanderDamageTaken, &st.SpellsCast, &st.CardsDrawn, &st.Attacks, &st.TokensCreated,
			&st.Mulligans, &cmdrs); err != nil {
			return Game{}, fmt.Errorf("gamestats: scan seat: %w", err)
		}
		st.Result, st.Place, st.EliminatedTurn, st.Conceded = Result(result), intPtr(place), intPtr(elimTurn), conceded == 1
		if err := json.Unmarshal([]byte(cmdrs), &st.Commanders); err != nil {
			return Game{}, fmt.Errorf("gamestats: stored commanders: %w", err)
		}
		g.Seats = append(g.Seats, st)
	}
	if err := rows.Err(); err != nil {
		return Game{}, fmt.Errorf("gamestats: seats: %w", err)
	}

	cards, err := s.db.QueryContext(ctx,
		`SELECT seat, name, scryfall_id, casts, first_cast_turn FROM seat_card_stats
		 WHERE game_id = ? ORDER BY seat, name`, id.String())
	if err != nil {
		return Game{}, fmt.Errorf("gamestats: get cards: %w", err)
	}
	defer func() { _ = cards.Close() }()
	bySeat := make(map[int]*Seat, len(g.Seats))
	for i := range g.Seats {
		bySeat[g.Seats[i].Seat] = &g.Seats[i]
	}
	for cards.Next() {
		var (
			seat int
			c    Card
			sid  sql.NullString
		)
		if err := cards.Scan(&seat, &c.Name, &sid, &c.Casts, &c.FirstCastTurn); err != nil {
			return Game{}, fmt.Errorf("gamestats: scan card: %w", err)
		}
		c.ScryfallID = sid.String
		if st := bySeat[seat]; st != nil {
			st.Cards = append(st.Cards, c)
		}
	}
	if err := cards.Err(); err != nil {
		return Game{}, fmt.Errorf("gamestats: cards: %w", err)
	}
	return g, nil
}

var _ Store = (*SQLStore)(nil)

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func intPtr(n sql.NullInt64) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int64)
	return &v
}

func nonNil(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}
