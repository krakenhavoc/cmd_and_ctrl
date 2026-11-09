package lobby

// pod_stats.go is ADR 0220 slice 5 = ADR 0217 slice 7 (#4880): a pod's
// stats and its leaderboard.
//
//   - GET /pods/{id}/stats?range=: each sharing member's record, the
//     pod's games (ended games where two or more sharing members sat),
//     how they were won, the first-turn effect, and head-to-head.
//   - GET /pods/{id}/leaderboard?range=: sharing members with at least
//     5 counted pod games, by win rate, then average place, then games.
//     Pod games only; there is no "all games" board (owner answer 7).
//
// Both are members only (a non-member, and a pod that does not exist,
// are the same 404) and need an account (a guest or the admin token is a
// 403). The rules every reader keeps (ADR 0217 §6, ADR 0220 §6):
//
//   - a member shows numbers only while their stats_prefs.share_pods is
//     on; joining a pod is not consent, and a member who is not sharing
//     is listed as "not sharing" with no numbers;
//   - a member with a block with the viewer, in either direction, is not
//     in the viewer's member list, leaderboard or head-to-head, and the
//     other rows keep their numbers and close their ranks up;
//   - the pod's own numbers (pod games, win kinds, turn order) are
//     counted over every sharing member, hidden ones too, so a block
//     hides a person without rewriting the group's history;
//   - a bot or agent seat is never a person's record, a deleted account
//     never appears, and no opponent outside the pod is named.
//
// Viewing needs no switch of the viewer's own: the share switch decides
// what a member shows to others, not what they may look at.

import (
	"context"
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/pods"
)

var errNoPodStats = httpError(http.StatusServiceUnavailable, "pod stats need the game database, and this server has none")

// podStatsMember is one row of the stats page's member list. The two
// tallies are absent while Sharing is false.
type podStatsMember struct {
	PodPerson
	Owner   bool             `json:"owner,omitempty"`
	You     bool             `json:"you,omitempty"`
	Sharing bool             `json:"sharing"`
	PodGame *gamestats.Tally `json:"pod_games,omitempty"`
	Overall *gamestats.Tally `json:"overall,omitempty"`
}

// podHeadToHead is a pair of sharing members over the games both sat in.
type podHeadToHead struct {
	A      PodPerson `json:"a"`
	B      PodPerson `json:"b"`
	Games  int       `json:"games"`
	AWins  int       `json:"a_wins"`
	BWins  int       `json:"b_wins"`
	Others int       `json:"others"`
}

// podStatsResponse is GET /pods/{id}/stats' body.
type podStatsResponse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Range gamestats.Range `json:"range"`
	// Games is the pod games in the range.
	Games      int                      `json:"games"`
	MinGames   int                      `json:"min_games"`
	Members    []podStatsMember         `json:"members"`
	WinKinds   []gamestats.PodWinKind   `json:"win_kinds"`
	TurnOrder  []gamestats.TurnOrderRow `json:"turn_order"`
	HeadToHead []podHeadToHead          `json:"head_to_head"`
}

// podBoardRow is one leaderboard row.
type podBoardRow struct {
	PodPerson
	You bool `json:"you,omitempty"`
	// Rank is absent on an unranked row.
	Rank int `json:"rank,omitempty"`
	gamestats.Tally
	WinRate float64 `json:"win_rate"`
	// Needs is how many more pod games an unranked row needs.
	Needs int `json:"needs,omitempty"`
}

// podLeaderboardResponse is GET /pods/{id}/leaderboard's body.
type podLeaderboardResponse struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Range    gamestats.Range `json:"range"`
	Games    int             `json:"games"`
	MinGames int             `json:"min_games"`
	Rows     []podBoardRow   `json:"rows"`
	// Unranked are sharing members with fewer than MinGames pod games.
	Unranked []podBoardRow `json:"unranked"`
}

// podStatsRead is the roster and the fold behind both routes.
type podStatsRead struct {
	roster pods.Roster
	rng    gamestats.Range
	fold   gamestats.PodStats
	me     string
}

func (c Config) readPodStats(ctx context.Context, me, pod string, rng gamestats.Range) (podStatsRead, error) {
	roster, err := c.Pods.Roster(ctx, me, pod)
	if err != nil {
		return podStatsRead{}, podError(c, err, "reading a pod's roster")
	}
	seats, err := c.PodStats.PodSeats(ctx, roster.Sharing, rng.Since(c.now()))
	if err != nil {
		c.logger().Error("reading a pod's stats failed", "err", err)
		return podStatsRead{}, httpError(http.StatusInternalServerError, "could not load the pod's stats; try again")
	}
	return podStatsRead{roster: roster, rng: rng, fold: gamestats.FoldPod(roster.Sharing, seats), me: me}, nil
}

// podStatsCaller is the signed-in member behind a pod stats route.
func podStatsCaller(c Config, r *http.Request) (string, gamestats.Range, error) {
	me, err := podCaller(c, r)
	if err != nil {
		return "", "", err
	}
	if c.PodStats == nil {
		return "", "", errNoPodStats
	}
	rng, ok := gamestats.ParseRange(r.URL.Query().Get("range"))
	if !ok {
		return "", "", httpError(http.StatusBadRequest, "range must be all, 90d or 30d")
	}
	return me, rng, nil
}

func (p podStatsRead) numbers(user string) (gamestats.PodMemberStats, bool) {
	for _, m := range p.fold.Members {
		if m.UserID == user {
			return m, true
		}
	}
	return gamestats.PodMemberStats{}, false
}

// getPodStats is GET /pods/{id}/stats.
func getPodStats(c Config, w http.ResponseWriter, r *http.Request) error {
	me, rng, err := podStatsCaller(c, r)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), statsTimeout)
	defer cancel()
	read, err := c.readPodStats(ctx, me, r.PathValue("id"), rng)
	if err != nil {
		return err
	}
	out := podStatsResponse{
		ID: read.roster.ID, Name: read.roster.Name, Range: rng, Games: read.fold.Games, MinGames: gamestats.PodMinGames,
		Members:  make([]podStatsMember, 0, len(read.roster.Visible)),
		WinKinds: read.fold.WinKinds, TurnOrder: read.fold.TurnOrder, HeadToHead: []podHeadToHead{},
	}
	people := map[string]PodPerson{}
	for _, m := range read.roster.Visible {
		person := PodPerson{UserID: m.UserID, DisplayName: m.DisplayName, AvatarURL: m.AvatarURL}
		row := podStatsMember{PodPerson: person, Owner: m.Owner, You: m.UserID == me, Sharing: m.Sharing}
		if n, ok := read.numbers(m.UserID); ok && m.Sharing {
			pod, overall := n.PodGame, n.Overall
			row.PodGame, row.Overall = &pod, &overall
			people[m.UserID] = person
		}
		out.Members = append(out.Members, row)
	}
	for _, h := range read.fold.HeadToHead {
		a, okA := people[h.A]
		b, okB := people[h.B]
		if !okA || !okB {
			continue
		}
		out.HeadToHead = append(out.HeadToHead, podHeadToHead{A: a, B: b, Games: h.Games, AWins: h.AWins, BWins: h.BWins, Others: h.Others})
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// getPodLeaderboard is GET /pods/{id}/leaderboard.
func getPodLeaderboard(c Config, w http.ResponseWriter, r *http.Request) error {
	me, rng, err := podStatsCaller(c, r)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), statsTimeout)
	defer cancel()
	read, err := c.readPodStats(ctx, me, r.PathValue("id"), rng)
	if err != nil {
		return err
	}
	person := map[string]pods.Member{}
	var in []gamestats.PodRank
	for _, m := range read.roster.Visible {
		n, ok := read.numbers(m.UserID)
		if !ok || !m.Sharing {
			continue
		}
		person[m.UserID] = m
		in = append(in, gamestats.PodRank{UserID: m.UserID, Tally: n.PodGame})
	}
	row := func(r gamestats.PodRank) podBoardRow {
		m := person[r.UserID]
		out := podBoardRow{
			PodPerson: PodPerson{UserID: m.UserID, DisplayName: m.DisplayName, AvatarURL: m.AvatarURL},
			You:       m.UserID == me, Tally: r.Tally,
		}
		if r.Tally.Games > 0 {
			out.WinRate = float64(r.Tally.Wins) / float64(r.Tally.Games)
		}
		return out
	}
	ranked, ranks, rest := gamestats.RankPod(in)
	out := podLeaderboardResponse{
		ID: read.roster.ID, Name: read.roster.Name, Range: rng, Games: read.fold.Games, MinGames: gamestats.PodMinGames,
		Rows: make([]podBoardRow, 0, len(ranked)), Unranked: make([]podBoardRow, 0, len(rest)),
	}
	for i, rk := range ranked {
		b := row(rk)
		b.Rank = ranks[i]
		out.Rows = append(out.Rows, b)
	}
	for _, rk := range rest {
		b := row(rk)
		b.Needs = gamestats.PodMinGames - rk.Tally.Games
		out.Unranked = append(out.Unranked, b)
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}
