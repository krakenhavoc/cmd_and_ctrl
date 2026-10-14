package gamestats

// user.go reads one person's own record (ADR 0165 §7, #3366). The
// person is never stored in the stats tables: a seat is theirs when
// seats.user_id is, through the join on (game_id, seat). Nothing here
// ever reads another person's seat.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// RecentMax is how many recent games UserStats lists.
const RecentMax = 20

// Tally is a count of results over the seats in a group.
type Tally struct {
	Games  int `json:"games"`
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
	// AvgPlace is the mean place over the group's seats that have one,
	// and nil when none do.
	AvgPlace *float64 `json:"avg_place"`
	placeSum int
	placed   int
}

func (t *Tally) add(result Result, place *int) {
	t.Games++
	switch result {
	case ResultWin:
		t.Wins++
	case ResultLoss:
		t.Losses++
	case ResultDraw:
		t.Draws++
	}
	if place != nil {
		t.placeSum += *place
		t.placed++
	}
}

func (t *Tally) finish() {
	if t.placed > 0 {
		v := float64(t.placeSum) / float64(t.placed)
		t.AvgPlace = &v
	}
}

// Group is a Tally under a name: a deck or a commander (pair).
type Group struct {
	Name string `json:"name"`
	Tally
}

// RecentGame is one finished game of the caller's.
type RecentGame struct {
	GameID     string   `json:"game_id"`
	Name       string   `json:"name"`
	EndedAt    int64    `json:"ended_at"` // Unix milliseconds
	Result     Result   `json:"result"`
	Place      *int     `json:"place"`
	Turns      int      `json:"turns"`
	Players    int      `json:"players"`
	Deck       string   `json:"deck,omitempty"`
	Commanders []string `json:"commanders"`
}

// TopCardsMax caps UserStats.TopCards.
const TopCardsMax = 10

// TopCard is a card the caller cast, and how often over their counted
// games (ADR 0179: the stats page's Buy links).
type TopCard struct {
	Name  string `json:"name"`
	Casts int    `json:"casts"`
}

// UserStats is the body of GET /me/stats.
//
// Definitions: a game counts when the caller's seat has a result of
// win, loss or draw. A table that ended without an outcome (result
// "none", for instance closed in the lobby) is left out of the record
// and the groups, and shows only under Unfinished. AvgPlace is over the
// counted seats that have a place. A deck is grouped by its saved deck
// id (its name when it has none); a commander is the seat's commander
// names joined, so a partner pair is one entry.
type UserStats struct {
	Tally
	Unfinished  int          `json:"unfinished"`
	ByDeck      []Group      `json:"by_deck"`
	ByCommander []Group      `json:"by_commander"`
	Recent      []RecentGame `json:"recent"`
	// TopCards is the caller's most-cast cards over the same counted
	// seats (at most TopCardsMax, most casts first, then by name).
	TopCards []TopCard `json:"top_cards"`
}

// ForUser implements Store.
func (s *SQLStore) ForUser(ctx context.Context, userID string) (UserStats, error) {
	out := UserStats{ByDeck: []Group{}, ByCommander: []Group{}, Recent: []RecentGame{}, TopCards: []TopCard{}}
	if userID == "" {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.game_id, g.name, COALESCE(g.ended_at, gs.computed_at), ss.result, ss.place,
		        gs.turns, s.deck_id, s.deck_name, ss.commanders,
		        (SELECT COUNT(*) FROM seat_stats x WHERE x.game_id = s.game_id)
		 FROM seats s
		 JOIN seat_stats ss ON ss.game_id = s.game_id AND ss.seat = s.seat
		 JOIN game_stats gs ON gs.game_id = s.game_id
		 JOIN games g ON g.id = s.game_id
		 WHERE s.user_id = ? AND s.bot_tier IS NULL AND s.agent_client IS NULL
		 ORDER BY COALESCE(g.ended_at, gs.computed_at) DESC, s.game_id`, userID)
	if err != nil {
		return UserStats{}, fmt.Errorf("gamestats: user stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	decks, cmdrs := map[string]*Group{}, map[string]*Group{}
	var deckOrder, cmdrOrder []string
	for rows.Next() {
		var (
			rg               RecentGame
			result, cmdrJSON string
			place            sql.NullInt64
			deckID, deckName sql.NullString
		)
		if err := rows.Scan(&rg.GameID, &rg.Name, &rg.EndedAt, &result, &place, &rg.Turns,
			&deckID, &deckName, &cmdrJSON, &rg.Players); err != nil {
			return UserStats{}, fmt.Errorf("gamestats: scan user stats: %w", err)
		}
		rg.Result, rg.Place, rg.Deck = Result(result), intPtr(place), deckName.String
		if err := json.Unmarshal([]byte(cmdrJSON), &rg.Commanders); err != nil {
			return UserStats{}, fmt.Errorf("gamestats: stored commanders: %w", err)
		}
		if rg.Commanders == nil {
			rg.Commanders = []string{}
		}
		if rg.Result == ResultNone {
			out.Unfinished++
			continue
		}
		out.Tally.add(rg.Result, rg.Place)
		bump := func(m map[string]*Group, order *[]string, key, name string) {
			grp := m[key]
			if grp == nil {
				// Rows run newest first, so the first name is the latest.
				grp = &Group{Name: name}
				m[key] = grp
				*order = append(*order, key)
			}
			grp.add(rg.Result, rg.Place)
		}
		if deckName.String != "" || deckID.String != "" {
			key := deckID.String
			if key == "" {
				key = "name:" + deckName.String
			}
			name := deckName.String
			if name == "" {
				name = "Unnamed deck"
			}
			bump(decks, &deckOrder, key, name)
		}
		if len(rg.Commanders) > 0 {
			name := strings.Join(rg.Commanders, " / ")
			bump(cmdrs, &cmdrOrder, name, name)
		}
		if len(out.Recent) < RecentMax {
			out.Recent = append(out.Recent, rg)
		}
	}
	if err := rows.Err(); err != nil {
		return UserStats{}, fmt.Errorf("gamestats: user stats: %w", err)
	}
	out.Tally.finish()
	out.ByDeck = finishGroups(decks, deckOrder)
	out.ByCommander = finishGroups(cmdrs, cmdrOrder)
	if out.TopCards, err = s.topCards(ctx, userID); err != nil {
		return UserStats{}, err
	}
	return out, nil
}

// topCards sums the casts of the caller's own human seats that have a
// result, the seats the record counts (ADR 0165 §7, §9).
func (s *SQLStore) topCards(ctx context.Context, userID string) ([]TopCard, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.name, SUM(c.casts) AS n
		 FROM seat_card_stats c
		 JOIN seats s ON s.game_id = c.game_id AND s.seat = c.seat
		 JOIN seat_stats ss ON ss.game_id = c.game_id AND ss.seat = c.seat
		 WHERE s.user_id = ? AND s.bot_tier IS NULL AND s.agent_client IS NULL
		   AND ss.result <> 'none'
		 GROUP BY c.name
		 ORDER BY n DESC, c.name
		 LIMIT ?`, userID, TopCardsMax)
	if err != nil {
		return nil, fmt.Errorf("gamestats: top cards: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := []TopCard{}
	for rows.Next() {
		var c TopCard
		if err := rows.Scan(&c.Name, &c.Casts); err != nil {
			return nil, fmt.Errorf("gamestats: scan top card: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gamestats: top cards: %w", err)
	}
	return out, nil
}

// finishGroups lists groups most-played first, ties in recency order.
func finishGroups(m map[string]*Group, order []string) []Group {
	out := make([]Group, 0, len(order))
	for _, k := range order {
		m[k].finish()
		out = append(out, *m[k])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Games > out[j].Games })
	return out
}
