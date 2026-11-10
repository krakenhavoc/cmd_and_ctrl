package gamestats

// pod.go is ADR 0217 slice 7 / ADR 0220 slice 5 (#4880): what a pod's
// stats read. The store hands back the counted seats of the pod's
// sharing members; Pod* below folds them into the pod's overview and its
// leaderboard. Nothing here knows about blocks: the caller names which
// members a viewer may see (ADR 0220 §6), and the pod's own numbers are
// computed over every sharing member either way.

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// PodMinGames is how many counted pod games a member needs to be ranked
// on the leaderboard (ADR 0217 owner answer 7, ADR 0220 §7).
const PodMinGames = 5

// PodSeat is one counted seat of a sharing member: a signed-in person's
// own seat (never a bot or agent seat), with the game-level columns the
// pod's numbers need.
type PodSeat struct {
	GameID  string
	UserID  string
	Seat    int
	EndedAt int64
	Result  Result
	Place   *int
	Players int
	// TurnPosition is the seat's place in turn order, nil for a game
	// recorded before version 2 or whose replay was not read.
	TurnPosition *int
	// WinKind is the game's, empty for a draw or an unread game.
	WinKind Kind
}

// PodStore reads the seats a pod's numbers stand on. *SQLStore is one.
type PodStore interface {
	// PodSeats is every counted seat of the listed users, newest game
	// first, that ended at or after since (Unix ms; 0 keeps all). A
	// seat with a bot tier or an agent client is never returned, and a
	// deleted account's seats are not either.
	PodSeats(ctx context.Context, userIDs []string, since int64) ([]PodSeat, error)
}

var _ PodStore = (*SQLStore)(nil)

// podSeatsMaxUsers caps one PodSeats call; a pod holds at most 12.
const podSeatsMaxUsers = 64

// PodSeats implements PodStore.
func (s *SQLStore) PodSeats(ctx context.Context, userIDs []string, since int64) ([]PodSeat, error) {
	if len(userIDs) == 0 {
		return []PodSeat{}, nil
	}
	if len(userIDs) > podSeatsMaxUsers {
		return nil, fmt.Errorf("gamestats: pod seats: %d users", len(userIDs))
	}
	marks := strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",")
	args := make([]any, 0, len(userIDs)+1)
	for _, u := range userIDs {
		args = append(args, u)
	}
	args = append(args, since)
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.game_id, s.user_id, s.seat, COALESCE(g.ended_at, gs.computed_at), ss.result, ss.place,
		        (SELECT COUNT(*) FROM seat_stats os WHERE os.game_id = s.game_id),
		        ss.turn_position, gs.win_kind
		 FROM seats s
		 JOIN users u ON u.id = s.user_id AND u.deleted_at IS NULL
		 JOIN seat_stats ss ON ss.game_id = s.game_id AND ss.seat = s.seat
		 JOIN game_stats gs ON gs.game_id = s.game_id
		 JOIN games g ON g.id = s.game_id
		 WHERE s.user_id IN (`+marks+`) AND s.bot_tier IS NULL AND s.agent_client IS NULL
		   AND ss.result <> 'none' AND COALESCE(g.ended_at, gs.computed_at) >= ?
		 ORDER BY COALESCE(g.ended_at, gs.computed_at) DESC, s.game_id, s.seat`, args...)
	if err != nil {
		return nil, fmt.Errorf("gamestats: pod seats: %w", err)
	}
	out := []PodSeat{}
	for rows.Next() {
		var (
			r             PodSeat
			result        string
			place, turnAt sql.NullInt64
			winKind       sql.NullString
		)
		if err := rows.Scan(&r.GameID, &r.UserID, &r.Seat, &r.EndedAt, &result, &place, &r.Players, &turnAt, &winKind); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("gamestats: scan pod seat: %w", err)
		}
		r.Result, r.Place, r.TurnPosition, r.WinKind = Result(result), intPtr(place), intPtr(turnAt), Kind(winKind.String)
		out = append(out, r)
	}
	if err := closeRows(rows); err != nil {
		return nil, fmt.Errorf("gamestats: pod seats: %w", err)
	}
	return out, nil
}

// PodMemberStats is one sharing member's numbers: their record in the
// pod's games, and their record over every game in the range.
type PodMemberStats struct {
	UserID  string
	PodGame Tally
	Overall Tally
}

// PodWinKind counts the pod games won in one way. A game counts once,
// whoever of the members sat in it.
type PodWinKind struct {
	Kind  Kind `json:"kind"`
	Games int  `json:"games"`
}

// PodHeadToHead is one pair of sharing members over the games both sat
// in: how often each won. Others is the games neither won (a third seat
// won, or a draw). A and B are user ids, A < B.
type PodHeadToHead struct {
	A, B   string
	Games  int
	AWins  int
	BWins  int
	Others int
}

// PodStats is the fold of a pod's seats.
type PodStats struct {
	// Games is the pod games: ended games where two or more sharing
	// members sat.
	Games int
	// Members is every sharing member's numbers, in the order given.
	Members []PodMemberStats
	// WinKinds counts the pod games by how they were won, most first.
	WinKinds []PodWinKind
	// TurnOrder is the first-turn effect over the pod games' seats of
	// sharing members at four-seat tables.
	TurnOrder []TurnOrderRow
	// HeadToHead has one row per pair of sharing members who sat
	// together at least once.
	HeadToHead []PodHeadToHead
}

// FoldPod folds seats (the sharing members' counted seats in a range)
// into the pod's numbers. sharing lists the sharing members' ids, in
// the order Members is to be returned in.
func FoldPod(sharing []string, seats []PodSeat) PodStats {
	byGame := map[string][]*PodSeat{}
	var order []string
	for i := range seats {
		s := &seats[i]
		if _, ok := byGame[s.GameID]; !ok {
			order = append(order, s.GameID)
		}
		byGame[s.GameID] = append(byGame[s.GameID], s)
	}
	idx := make(map[string]int, len(sharing))
	out := PodStats{Members: make([]PodMemberStats, len(sharing)), WinKinds: []PodWinKind{}, TurnOrder: []TurnOrderRow{}, HeadToHead: []PodHeadToHead{}}
	for i, u := range sharing {
		idx[u] = i
		out.Members[i].UserID = u
	}
	for i := range seats {
		s := &seats[i]
		if m, ok := idx[s.UserID]; ok {
			out.Members[m].Overall.add(s.Result, s.Place)
		}
	}
	kinds := map[Kind]int{}
	type posAcc struct {
		games, wins, placed, placeSum int
	}
	pos := map[int]*posAcc{}
	pairs := map[[2]string]*PodHeadToHead{}
	for _, gid := range order {
		g := byGame[gid]
		people := map[string]*PodSeat{}
		for _, s := range g {
			if _, ok := idx[s.UserID]; ok {
				people[s.UserID] = s
			}
		}
		if len(people) < 2 {
			continue
		}
		out.Games++
		if g[0].WinKind != "" {
			kinds[g[0].WinKind]++
		}
		ids := make([]string, 0, len(people))
		for u, s := range people {
			ids = append(ids, u)
			out.Members[idx[u]].PodGame.add(s.Result, s.Place)
			if s.Players == 4 && s.TurnPosition != nil && *s.TurnPosition >= 1 && *s.TurnPosition <= 4 {
				a := pos[*s.TurnPosition]
				if a == nil {
					a = &posAcc{}
					pos[*s.TurnPosition] = a
				}
				a.games++
				if s.Result == ResultWin {
					a.wins++
				}
				if s.Place != nil {
					a.placed++
					a.placeSum += *s.Place
				}
			}
		}
		sort.Strings(ids)
		for i := range ids {
			for j := i + 1; j < len(ids); j++ {
				key := [2]string{ids[i], ids[j]}
				h := pairs[key]
				if h == nil {
					h = &PodHeadToHead{A: key[0], B: key[1]}
					pairs[key] = h
				}
				h.Games++
				switch {
				case people[key[0]].Result == ResultWin:
					h.AWins++
				case people[key[1]].Result == ResultWin:
					h.BWins++
				default:
					h.Others++
				}
			}
		}
	}
	for i := range out.Members {
		out.Members[i].PodGame.finish()
		out.Members[i].Overall.finish()
	}
	for k, n := range kinds {
		out.WinKinds = append(out.WinKinds, PodWinKind{Kind: k, Games: n})
	}
	sort.Slice(out.WinKinds, func(i, j int) bool {
		if out.WinKinds[i].Games != out.WinKinds[j].Games {
			return out.WinKinds[i].Games > out.WinKinds[j].Games
		}
		return out.WinKinds[i].Kind < out.WinKinds[j].Kind
	})
	for p := 1; p <= 4; p++ {
		a := pos[p]
		if a == nil {
			continue
		}
		row := TurnOrderRow{Position: p, Wins: RateOf(a.wins, a.games)}
		row.AvgPlace = mean(a.placeSum, a.placed)
		out.TurnOrder = append(out.TurnOrder, row)
	}
	for _, h := range pairs {
		out.HeadToHead = append(out.HeadToHead, *h)
	}
	sort.Slice(out.HeadToHead, func(i, j int) bool {
		a, b := out.HeadToHead[i], out.HeadToHead[j]
		if a.Games != b.Games {
			return a.Games > b.Games
		}
		if a.A != b.A {
			return a.A < b.A
		}
		return a.B < b.B
	})
	return out
}

// PodRank is one leaderboard row's inputs.
type PodRank struct {
	UserID string
	Tally  Tally
}

// RankPod orders the members who have at least PodMinGames counted pod
// games by win rate, then average place (lower first), then games (more
// first), then user id. It returns the ranked rows and, apart, the rest:
// members with fewer games, the closest to the minimum first, so a row
// can say how many more they need. Equal rows share a rank (1-based, then
// skipping, as a sports table does).
func RankPod(in []PodRank) (ranked []PodRank, ranks []int, rest []PodRank) {
	for _, r := range in {
		if r.Tally.Games >= PodMinGames {
			ranked = append(ranked, r)
		} else {
			rest = append(rest, r)
		}
	}
	winRate := func(t Tally) float64 { return float64(t.Wins) / float64(t.Games) }
	place := func(t Tally) float64 {
		if t.AvgPlace == nil {
			return 99
		}
		return *t.AvgPlace
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		a, b := ranked[i].Tally, ranked[j].Tally
		if wa, wb := winRate(a), winRate(b); wa != wb {
			return wa > wb
		}
		if pa, pb := place(a), place(b); pa != pb {
			return pa < pb
		}
		if a.Games != b.Games {
			return a.Games > b.Games
		}
		return ranked[i].UserID < ranked[j].UserID
	})
	ranks = make([]int, len(ranked))
	for i := range ranked {
		ranks[i] = i + 1
		if i > 0 {
			a, b := ranked[i-1].Tally, ranked[i].Tally
			if winRate(a) == winRate(b) && place(a) == place(b) && a.Games == b.Games {
				ranks[i] = ranks[i-1]
			}
		}
	}
	sort.SliceStable(rest, func(i, j int) bool {
		if rest[i].Tally.Games != rest[j].Tally.Games {
			return rest[i].Tally.Games > rest[j].Tally.Games
		}
		return rest[i].UserID < rest[j].UserID
	})
	return ranked, ranks, rest
}
