package gamestats

// admin.go reads the whole table's statistics for an operator (ADR 0165
// §8, #3425): GET /admin/stats. It reuses the user view's definitions:
// only a seat with a win, loss or draw counts, average place is over the
// counted seats that have a place, and a commander is the seat's
// commander names joined with " / ".

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DeletedPlayersName labels the one row that holds every deleted account.
const DeletedPlayersName = "Deleted players"

// AdminListMax caps every list in an Overview.
const AdminListMax = 25

// Totals counts games with a record, and the games that should have one
// and do not.
type Totals struct {
	Games         int `json:"games"`
	Live          int `json:"live"`
	Replay        int `json:"replay"`
	LogComplete   int `json:"log_complete"`
	LogIncomplete int `json:"log_incomplete"`
	// Unrecorded counts ended, stored games with no record: what the
	// backfill (cmd/statsbackfill) would fill in.
	Unrecorded int `json:"unrecorded"`
}

// AdminGame is one recent game with a record.
type AdminGame struct {
	GameID  string `json:"game_id"`
	Name    string `json:"name"`
	EndedAt int64  `json:"ended_at"` // Unix milliseconds
	Turns   int    `json:"turns"`
	Rounds  int    `json:"rounds"`
	Source  string `json:"source"`
}

// PersonRecord is one person's (or guest label's, or bot tier's) record.
// Kind is "account", "guest", "bot", "agent" or "deleted". Every deleted
// account (ADR 0149) is folded into the one "deleted" row, so a deleted
// person's history cannot be picked out (ADR 0165 §9).
type PersonRecord struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Tally
}

// CardCasts is a card's cast count across all games.
type CardCasts struct {
	Name  string `json:"name"`
	Casts int    `json:"casts"`
	Games int    `json:"games"`
}

// Overview is the body of GET /admin/stats.
type Overview struct {
	Totals          Totals         `json:"totals"`
	Recent          []AdminGame    `json:"recent"`
	People          []PersonRecord `json:"people"`
	CommandersGames []Group        `json:"commanders_by_games"`
	CommandersWins  []Group        `json:"commanders_by_wins"`
	Cards           []CardCasts    `json:"cards"`
}

// EmptyOverview is an Overview with no data and empty (not null) lists.
func EmptyOverview() Overview {
	return Overview{Recent: []AdminGame{}, People: []PersonRecord{}, CommandersGames: []Group{},
		CommandersWins: []Group{}, Cards: []CardCasts{}}
}

// Overview implements Store.
func (s *SQLStore) Overview(ctx context.Context) (Overview, error) {
	out := EmptyOverview()
	fail := func(what string, err error) (Overview, error) {
		return EmptyOverview(), fmt.Errorf("gamestats: overview %s: %w", what, err)
	}

	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*),
		        COALESCE(SUM(source = 'live'), 0), COALESCE(SUM(source = 'replay'), 0),
		        COALESCE(SUM(log_complete = 1), 0), COALESCE(SUM(log_complete = 0), 0)
		 FROM game_stats`).Scan(&out.Totals.Games, &out.Totals.Live, &out.Totals.Replay,
		&out.Totals.LogComplete, &out.Totals.LogIncomplete)
	if err != nil {
		return fail("totals", err)
	}
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM games g
		 WHERE g.state = 'ended' AND NOT EXISTS (SELECT 1 FROM game_stats gs WHERE gs.game_id = g.id)`).
		Scan(&out.Totals.Unrecorded); err != nil {
		return fail("unrecorded", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT g.id, g.name, COALESCE(g.ended_at, gs.computed_at), gs.turns, gs.rounds, gs.source
		 FROM game_stats gs JOIN games g ON g.id = gs.game_id
		 ORDER BY COALESCE(g.ended_at, gs.computed_at) DESC, g.id LIMIT ?`, AdminListMax)
	if err != nil {
		return fail("recent", err)
	}
	for rows.Next() {
		var g AdminGame
		if err := rows.Scan(&g.GameID, &g.Name, &g.EndedAt, &g.Turns, &g.Rounds, &g.Source); err != nil {
			_ = rows.Close()
			return fail("recent", err)
		}
		out.Recent = append(out.Recent, g)
	}
	if err := closeRows(rows); err != nil {
		return fail("recent", err)
	}

	// Every counted seat, with the person it belongs to.
	rows, err = s.db.QueryContext(ctx,
		`SELECT s.user_id, u.display_name, u.deleted_at, s.guest_name, s.bot_tier, s.agent_client,
		        ss.result, ss.place, ss.commanders
		 FROM seat_stats ss
		 JOIN seats s ON s.game_id = ss.game_id AND s.seat = ss.seat
		 LEFT JOIN users u ON u.id = s.user_id
		 WHERE ss.result IN ('win', 'loss', 'draw')`)
	if err != nil {
		return fail("seats", err)
	}
	people, cmdrs := map[string]*PersonRecord{}, map[string]*Group{}
	for rows.Next() {
		var (
			userID, display, guest, tier, agent sql.NullString
			deletedAt                           sql.NullInt64
			result, cmdrJSON                    string
			place                               sql.NullInt64
		)
		if err := rows.Scan(&userID, &display, &deletedAt, &guest, &tier, &agent, &result, &place, &cmdrJSON); err != nil {
			_ = rows.Close()
			return fail("seats", err)
		}
		p := intPtr(place)
		var key, name, kind string
		switch {
		case agent.String != "":
			key, name, kind = "a:"+agent.String, "AI agent ("+agent.String+")", "agent"
		case tier.String != "":
			key, name, kind = "b:"+tier.String, "Bot ("+tier.String+")", "bot"
		case userID.String != "" && deletedAt.Valid:
			key, name, kind = "deleted", DeletedPlayersName, "deleted"
		case userID.String != "":
			key, name, kind = "u:"+userID.String, display.String, "account"
			if name == "" {
				name = "Unknown player"
			}
		default:
			name = guest.String
			if name == "" {
				name = "Guest"
			}
			key, kind = "g:"+name, "guest"
		}
		rec := people[key]
		if rec == nil {
			rec = &PersonRecord{Name: name, Kind: kind}
			people[key] = rec
		}
		rec.add(Result(result), p)

		var names []string
		if err := json.Unmarshal([]byte(cmdrJSON), &names); err != nil {
			_ = rows.Close()
			return fail("commanders", err)
		}
		// Commanders count human seats only.
		if len(names) > 0 && tier.String == "" && agent.String == "" {
			cn := strings.Join(names, " / ")
			grp := cmdrs[cn]
			if grp == nil {
				grp = &Group{Name: cn}
				cmdrs[cn] = grp
			}
			grp.add(Result(result), p)
		}
	}
	if err := closeRows(rows); err != nil {
		return fail("seats", err)
	}

	pl := make([]PersonRecord, 0, len(people))
	for _, r := range people {
		r.finish()
		pl = append(pl, *r)
	}
	sort.Slice(pl, func(i, j int) bool {
		if pl[i].Games != pl[j].Games {
			return pl[i].Games > pl[j].Games
		}
		return pl[i].Name < pl[j].Name
	})
	out.People = capList(pl)

	all := make([]Group, 0, len(cmdrs))
	for _, g := range cmdrs {
		g.finish()
		all = append(all, *g)
	}
	ranked := func(count func(Group) int) []Group {
		l := append([]Group(nil), all...)
		sort.Slice(l, func(i, j int) bool {
			if ci, cj := count(l[i]), count(l[j]); ci != cj {
				return ci > cj
			}
			return l[i].Name < l[j].Name
		})
		return capList(l)
	}
	out.CommandersGames = ranked(func(g Group) int { return g.Games })
	out.CommandersWins = ranked(func(g Group) int { return g.Wins })

	rows, err = s.db.QueryContext(ctx,
		`SELECT c.name, SUM(c.casts), COUNT(DISTINCT c.game_id) FROM seat_card_stats c
		 JOIN seats s ON s.game_id = c.game_id AND s.seat = c.seat
		 WHERE s.bot_tier IS NULL AND s.agent_client IS NULL
		 GROUP BY c.name ORDER BY SUM(c.casts) DESC, c.name LIMIT ?`, AdminListMax)
	if err != nil {
		return fail("cards", err)
	}
	for rows.Next() {
		var c CardCasts
		if err := rows.Scan(&c.Name, &c.Casts, &c.Games); err != nil {
			_ = rows.Close()
			return fail("cards", err)
		}
		out.Cards = append(out.Cards, c)
	}
	if err := closeRows(rows); err != nil {
		return fail("cards", err)
	}
	return out, nil
}

func closeRows(rows *sql.Rows) error {
	err := rows.Err()
	if cerr := rows.Close(); err == nil {
		err = cerr
	}
	return err
}

func capList[T any](l []T) []T {
	if len(l) > AdminListMax {
		l = l[:AdminListMax]
	}
	if l == nil {
		return []T{}
	}
	return l
}

// Overview implements Store; there is nothing to show.
func (NoStore) Overview(context.Context) (Overview, error) { return EmptyOverview(), nil }
