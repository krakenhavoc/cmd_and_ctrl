package gamestats

// store_v2.go writes and reads migration 0029's per-seat tables (ADR
// 0217 §1): seat_turn_stats, seat_target_stats and seat_card_seen.

import (
	"context"
	"database/sql"
	"fmt"
)

// v2Stmts are the per-row inserts, prepared once per Put.
type v2Stmts struct {
	turn, target, seen *sql.Stmt
}

func prepareV2(ctx context.Context, tx *sql.Tx) (*v2Stmts, error) {
	var (
		s   v2Stmts
		err error
	)
	if s.turn, err = tx.PrepareContext(ctx,
		`INSERT INTO seat_turn_stats (game_id, seat, own_turn, turn, land_played, land_drops, lands_controlled,
		   mana_sources, hand_size, lands_in_hand, spells_in_hand, missed, screw, flood, colour_screw,
		   attackers, eligible, damage_to_players, spells_cast)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`); err != nil {
		return nil, fmt.Errorf("gamestats: prepare turns: %w", err)
	}
	if s.target, err = tx.PrepareContext(ctx,
		`INSERT INTO seat_target_stats (game_id, seat, target_seat, attacks, damage, combat_damage, commander_damage)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`); err != nil {
		s.close()
		return nil, fmt.Errorf("gamestats: prepare targets: %w", err)
	}
	if s.seen, err = tx.PrepareContext(ctx,
		`INSERT INTO seat_card_seen (game_id, seat, name, copies, in_opening_hand, first_seen_turn, held_turns,
		   affordable_turns, "cast")
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`); err != nil {
		s.close()
		return nil, fmt.Errorf("gamestats: prepare cards seen: %w", err)
	}
	return &s, nil
}

func (s *v2Stmts) close() {
	for _, st := range []*sql.Stmt{s.turn, s.target, s.seen} {
		if st != nil {
			_ = st.Close()
		}
	}
}

// put writes one seat's version-2 rows.
func (s *v2Stmts) put(ctx context.Context, id string, st Seat) error {
	for _, t := range st.Targets {
		if _, err := s.target.ExecContext(ctx, id, st.Seat, t.TargetSeat, t.Attacks, t.Damage, t.CombatDamage,
			t.CommanderDamage); err != nil {
			return fmt.Errorf("target %d: %w", t.TargetSeat, err)
		}
	}
	if st.Play == nil {
		return nil
	}
	for _, t := range st.Play.Turns {
		if _, err := s.turn.ExecContext(ctx, id, st.Seat, t.OwnTurn, t.Turn, t.LandPlayed, t.LandDrops,
			t.LandsControlled, t.ManaSources, t.HandSize, t.LandsInHand, t.SpellsInHand, boolInt(t.Missed),
			boolInt(t.Screw), boolInt(t.Flood), boolInt(t.ColourScrew), t.Attackers, t.Eligible,
			t.DamageToPlayers, t.SpellsCast); err != nil {
			return fmt.Errorf("own turn %d: %w", t.OwnTurn, err)
		}
	}
	for _, c := range st.Play.Seen {
		if _, err := s.seen.ExecContext(ctx, id, st.Seat, c.Name, c.Copies, boolInt(c.InOpeningHand),
			c.FirstSeenTurn, c.HeldTurns, c.AffordableTurns, boolInt(c.Cast)); err != nil {
			return fmt.Errorf("card seen %q: %w", c.Name, err)
		}
	}
	return nil
}

// getV2 reads a game's version-2 rows onto its seats. A version-1
// record has none, and its seats keep nil Targets.
func (s *SQLStore) getV2(ctx context.Context, id string, version int, bySeat map[int]*Seat) error {
	if version < 2 {
		return nil
	}
	for _, st := range bySeat {
		st.Targets = []TargetStat{}
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT seat, target_seat, attacks, damage, combat_damage, commander_damage FROM seat_target_stats
		 WHERE game_id = ? ORDER BY seat, target_seat`, id)
	if err != nil {
		return fmt.Errorf("gamestats: get targets: %w", err)
	}
	for rows.Next() {
		var (
			seat int
			t    TargetStat
		)
		if err := rows.Scan(&seat, &t.TargetSeat, &t.Attacks, &t.Damage, &t.CombatDamage, &t.CommanderDamage); err != nil {
			_ = rows.Close()
			return fmt.Errorf("gamestats: scan target: %w", err)
		}
		if st := bySeat[seat]; st != nil {
			st.Targets = append(st.Targets, t)
		}
	}
	if err := closeRows(rows); err != nil {
		return fmt.Errorf("gamestats: targets: %w", err)
	}

	rows, err = s.db.QueryContext(ctx,
		`SELECT seat, own_turn, turn, land_played, land_drops, lands_controlled, mana_sources, hand_size,
		   lands_in_hand, spells_in_hand, missed, screw, flood, colour_screw, attackers, eligible,
		   damage_to_players, spells_cast
		 FROM seat_turn_stats WHERE game_id = ? ORDER BY seat, own_turn`, id)
	if err != nil {
		return fmt.Errorf("gamestats: get turns: %w", err)
	}
	for rows.Next() {
		var (
			seat                             int
			t                                TurnStat
			missed, screw, flood, colourScrw int
		)
		if err := rows.Scan(&seat, &t.OwnTurn, &t.Turn, &t.LandPlayed, &t.LandDrops, &t.LandsControlled,
			&t.ManaSources, &t.HandSize, &t.LandsInHand, &t.SpellsInHand, &missed, &screw, &flood, &colourScrw,
			&t.Attackers, &t.Eligible, &t.DamageToPlayers, &t.SpellsCast); err != nil {
			_ = rows.Close()
			return fmt.Errorf("gamestats: scan turn: %w", err)
		}
		t.Missed, t.Screw, t.Flood, t.ColourScrew = missed == 1, screw == 1, flood == 1, colourScrw == 1
		if st := bySeat[seat]; st != nil && st.Play != nil {
			st.Play.Turns = append(st.Play.Turns, t)
		}
	}
	if err := closeRows(rows); err != nil {
		return fmt.Errorf("gamestats: turns: %w", err)
	}

	rows, err = s.db.QueryContext(ctx,
		`SELECT seat, name, copies, in_opening_hand, first_seen_turn, held_turns, affordable_turns, "cast"
		 FROM seat_card_seen WHERE game_id = ? ORDER BY seat, name`, id)
	if err != nil {
		return fmt.Errorf("gamestats: get cards seen: %w", err)
	}
	for rows.Next() {
		var (
			seat          int
			c             SeenCard
			opening, cast int
		)
		if err := rows.Scan(&seat, &c.Name, &c.Copies, &opening, &c.FirstSeenTurn, &c.HeldTurns,
			&c.AffordableTurns, &cast); err != nil {
			_ = rows.Close()
			return fmt.Errorf("gamestats: scan card seen: %w", err)
		}
		c.InOpeningHand, c.Cast = opening == 1, cast == 1
		if st := bySeat[seat]; st != nil && st.Play != nil {
			st.Play.Seen = append(st.Play.Seen, c)
		}
	}
	if err := closeRows(rows); err != nil {
		return fmt.Errorf("gamestats: cards seen: %w", err)
	}
	return nil
}
