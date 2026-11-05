package gamestats

// advanced_store.go reads one person's version-2 record for ADR 0217's
// read routes (slice 5 of #4880): GET /me/stats/advanced and GET
// /me/stats/decks/{deck_id}. As user.go: a seat is the person's when
// seats.user_id is, and a bot or agent seat is never a person's record
// (ADR 0165 §9). Nothing here returns another seat's person: the other
// seats are read only as counts (table makeup) and as seat numbers
// (seat_target_stats).

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// SeatRow is one of the person's counted seats (a result of win, loss
// or draw), with its version-2 columns. A pointer is nil when the
// column is NULL: a version-1 record, or a game whose replay was not
// read (frames_complete 0).
type SeatRow struct {
	GameID string
	Seat   int
	// EndedAt is Unix milliseconds: the game's end, or the record's
	// computation when the game row has none. StartedAt is 0 when the
	// game row has no start.
	EndedAt   int64
	StartedAt int64
	Result    Result
	Place     *int
	// Players is the seats in the game's record.
	Players    int
	Turns      int
	Rounds     int
	Mulligans  int
	DeckID     string
	DeckName   string
	Commanders []string
	// OtherBots and OtherPeople count the game's other seats: a bot is
	// a seat with a bot tier or an agent client, anyone else a person
	// (guests included). ADR 0217 §3's table makeup.
	OtherBots   int
	OtherPeople int
	// StandIn: a stand-in bot played some of this seat's turns (ADR
	// 0177). Its turns count as the person's.
	StandIn bool

	StatsVersion   int
	FramesComplete *bool
	// WinKind and ComboRule are the game's: how it was won.
	WinKind   Kind
	ComboRule ComboRule
	// OutKind is how this seat left, empty when it did not.
	OutKind      Kind
	TurnPosition *int
	DeckHash     string

	DamageToPlayers        int
	OwnTurns               *int
	LandDropsHit           *int
	LandDropsMissed        *int
	ScrewGame              *bool
	FloodGame              *bool
	ColourScrewGame        *bool
	CreaturesEligible      *int
	CreaturesAttacked      *int
	CommanderCasts         *int
	CommanderFirstCastTurn *int

	TurnRows []TurnRow
	Targets  []TargetRow
	Cast     []CastRow
	Seen     []SeenRow
}

// TurnRow is one seat_turn_stats row.
type TurnRow struct {
	OwnTurn, Turn, LandPlayed, LandDrops, LandsControlled int
	Missed                                                bool
	Attackers, Eligible, DamageToPlayers                  int
}

// TargetRow is one seat_target_stats row: what the seat did to one
// opponent seat, never named.
type TargetRow struct {
	TargetSeat, Attacks, Damage int
}

// CastRow is one seat_card_stats row.
type CastRow struct {
	Name  string
	Casts int
}

// SeenRow is one seat_card_seen row.
type SeenRow struct {
	Name            string
	AffordableTurns int
	Cast            bool
}

// AdvancedStore reads a person's version-2 record. *SQLStore is one.
type AdvancedStore interface {
	// AdvancedSeats is every counted seat of userID's own (no bot or
	// agent seat), newest first, that keep accepts (nil keeps every
	// one), with its turn, target and card rows.
	AdvancedSeats(ctx context.Context, userID string, keep func(*SeatRow) bool) ([]SeatRow, error)
}

var _ AdvancedStore = (*SQLStore)(nil)

// mineCTE is the person's counted seats: the same rule as ForUser.
// Each child query joins it, so they read through seats_user_id and
// the child tables' (game_id, seat) primary keys, never a whole table.
const mineCTE = `WITH mine AS (
	SELECT s.game_id, s.seat FROM seats s
	JOIN seat_stats ss ON ss.game_id = s.game_id AND ss.seat = s.seat
	WHERE s.user_id = ? AND s.bot_tier IS NULL AND s.agent_client IS NULL AND ss.result <> 'none')
`

// advancedSeatsQuery is the person's counted seats with their
// version-2 columns, newest first. It reads through seats_user_id, and
// the makeup counts through the other seats' primary keys.
const advancedSeatsQuery = `SELECT s.game_id, s.seat, COALESCE(g.ended_at, gs.computed_at), COALESCE(g.started_at, 0),
        ss.result, ss.place, gs.turns, gs.rounds, ss.mulligans, s.deck_id, s.deck_name, ss.commanders,
        s.stand_in IS NOT NULL,
        (SELECT COUNT(*) FROM seat_stats os JOIN seats o ON o.game_id = os.game_id AND o.seat = os.seat
          WHERE os.game_id = s.game_id AND os.seat <> s.seat AND (o.bot_tier IS NOT NULL OR o.agent_client IS NOT NULL)),
        (SELECT COUNT(*) FROM seat_stats os WHERE os.game_id = s.game_id),
        gs.stats_version, gs.frames_complete, gs.win_kind, gs.combo_rule,
        ss.out_kind, ss.turn_position, ss.deck_hash, ss.damage_to_players, ss.own_turns,
        ss.land_drops_hit, ss.land_drops_missed, ss.screw_game, ss.flood_game, ss.colour_screw_game,
        ss.creatures_eligible, ss.creatures_attacked, ss.commander_casts, ss.commander_first_cast_turn
 FROM seats s
 JOIN seat_stats ss ON ss.game_id = s.game_id AND ss.seat = s.seat
 JOIN game_stats gs ON gs.game_id = s.game_id
 JOIN games g ON g.id = s.game_id
 WHERE s.user_id = ? AND s.bot_tier IS NULL AND s.agent_client IS NULL AND ss.result <> 'none'
 ORDER BY COALESCE(g.ended_at, gs.computed_at) DESC, s.game_id`

// AdvancedSeats implements AdvancedStore.
func (s *SQLStore) AdvancedSeats(ctx context.Context, userID string, keep func(*SeatRow) bool) ([]SeatRow, error) {
	if userID == "" {
		return []SeatRow{}, nil
	}
	rows, err := s.db.QueryContext(ctx, advancedSeatsQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("gamestats: advanced seats: %w", err)
	}
	out := []SeatRow{}
	for rows.Next() {
		var (
			r                                           SeatRow
			result, cmdrJSON                            string
			place, framesComplete, turnPos              sql.NullInt64
			ownTurns, hit, missed, screw, flood, colour sql.NullInt64
			eligible, attacked, casts, firstCast        sql.NullInt64
			deckID, deckName, winKind, comboRule        sql.NullString
			outKind, deckHash                           sql.NullString
			players                                     int
		)
		if err := rows.Scan(&r.GameID, &r.Seat, &r.EndedAt, &r.StartedAt, &result, &place, &r.Turns, &r.Rounds,
			&r.Mulligans, &deckID, &deckName, &cmdrJSON, &r.StandIn, &r.OtherBots, &players,
			&r.StatsVersion, &framesComplete, &winKind, &comboRule, &outKind, &turnPos, &deckHash,
			&r.DamageToPlayers, &ownTurns, &hit, &missed, &screw, &flood, &colour,
			&eligible, &attacked, &casts, &firstCast); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("gamestats: scan advanced seat: %w", err)
		}
		r.Result, r.Place, r.Players = Result(result), intPtr(place), players
		r.OtherPeople = players - 1 - r.OtherBots
		r.DeckID, r.DeckName = deckID.String, deckName.String
		if err := json.Unmarshal([]byte(cmdrJSON), &r.Commanders); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("gamestats: stored commanders: %w", err)
		}
		r.Commanders = nonNil(r.Commanders)
		r.FramesComplete = boolPtr(framesComplete)
		r.WinKind, r.ComboRule, r.OutKind = Kind(winKind.String), ComboRule(comboRule.String), Kind(outKind.String)
		r.TurnPosition, r.DeckHash = intPtr(turnPos), deckHash.String
		r.OwnTurns, r.LandDropsHit, r.LandDropsMissed = intPtr(ownTurns), intPtr(hit), intPtr(missed)
		r.ScrewGame, r.FloodGame, r.ColourScrewGame = boolPtr(screw), boolPtr(flood), boolPtr(colour)
		r.CreaturesEligible, r.CreaturesAttacked = intPtr(eligible), intPtr(attacked)
		r.CommanderCasts, r.CommanderFirstCastTurn = intPtr(casts), intPtr(firstCast)
		if keep == nil || keep(&r) {
			out = append(out, r)
		}
	}
	if err := closeRows(rows); err != nil {
		return nil, fmt.Errorf("gamestats: advanced seats: %w", err)
	}
	if len(out) == 0 {
		return out, nil
	}
	at := make(map[[2]any]int, len(out))
	for i := range out {
		at[[2]any{out[i].GameID, out[i].Seat}] = i
	}
	// each runs one child query over the person's seats and hands every
	// row that belongs to a kept seat to add.
	each := func(what, q string, scan func(*sql.Rows, *SeatRow) error) error {
		rows, err := s.db.QueryContext(ctx, mineCTE+q, userID)
		if err != nil {
			return fmt.Errorf("gamestats: advanced %s: %w", what, err)
		}
		for rows.Next() {
			// scan reads the key into holder.GameID and holder.Seat and
			// the row into the matching child list.
			var holder SeatRow
			if err := scan(rows, &holder); err != nil {
				_ = rows.Close()
				return fmt.Errorf("gamestats: scan advanced %s: %w", what, err)
			}
			i, ok := at[[2]any{holder.GameID, holder.Seat}]
			if !ok {
				continue
			}
			r := &out[i]
			r.TurnRows = append(r.TurnRows, holder.TurnRows...)
			r.Targets = append(r.Targets, holder.Targets...)
			r.Cast = append(r.Cast, holder.Cast...)
			r.Seen = append(r.Seen, holder.Seen...)
		}
		if err := closeRows(rows); err != nil {
			return fmt.Errorf("gamestats: advanced %s: %w", what, err)
		}
		return nil
	}
	if err := each("turns",
		`SELECT t.game_id, t.seat, t.own_turn, t.turn, t.land_played, t.land_drops, t.lands_controlled,
		        t.missed, t.attackers, t.eligible, t.damage_to_players
		 FROM mine JOIN seat_turn_stats t ON t.game_id = mine.game_id AND t.seat = mine.seat
		 ORDER BY t.game_id, t.seat, t.own_turn`,
		func(rows *sql.Rows, h *SeatRow) error {
			var t TurnRow
			err := rows.Scan(&h.GameID, &h.Seat, &t.OwnTurn, &t.Turn, &t.LandPlayed, &t.LandDrops,
				&t.LandsControlled, &t.Missed, &t.Attackers, &t.Eligible, &t.DamageToPlayers)
			h.TurnRows = []TurnRow{t}
			return err
		}); err != nil {
		return nil, err
	}
	if err := each("targets",
		`SELECT t.game_id, t.seat, t.target_seat, t.attacks, t.damage
		 FROM mine JOIN seat_target_stats t ON t.game_id = mine.game_id AND t.seat = mine.seat
		 ORDER BY t.game_id, t.seat, t.target_seat`,
		func(rows *sql.Rows, h *SeatRow) error {
			var t TargetRow
			err := rows.Scan(&h.GameID, &h.Seat, &t.TargetSeat, &t.Attacks, &t.Damage)
			h.Targets = []TargetRow{t}
			return err
		}); err != nil {
		return nil, err
	}
	if err := each("casts",
		`SELECT c.game_id, c.seat, c.name, c.casts
		 FROM mine JOIN seat_card_stats c ON c.game_id = mine.game_id AND c.seat = mine.seat
		 ORDER BY c.game_id, c.seat, c.name`,
		func(rows *sql.Rows, h *SeatRow) error {
			var c CastRow
			err := rows.Scan(&h.GameID, &h.Seat, &c.Name, &c.Casts)
			h.Cast = []CastRow{c}
			return err
		}); err != nil {
		return nil, err
	}
	if err := each("cards seen",
		`SELECT c.game_id, c.seat, c.name, c.affordable_turns, c."cast"
		 FROM mine JOIN seat_card_seen c ON c.game_id = mine.game_id AND c.seat = mine.seat
		 ORDER BY c.game_id, c.seat, c.name`,
		func(rows *sql.Rows, h *SeatRow) error {
			var c SeenRow
			err := rows.Scan(&h.GameID, &h.Seat, &c.Name, &c.AffordableTurns, &c.Cast)
			h.Seen = []SeenRow{c}
			return err
		}); err != nil {
		return nil, err
	}
	return out, nil
}

func boolPtr(n sql.NullInt64) *bool {
	if !n.Valid {
		return nil
	}
	v := n.Int64 != 0
	return &v
}
