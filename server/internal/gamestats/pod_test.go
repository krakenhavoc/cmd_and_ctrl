package gamestats

import "testing"

func podSeat(game, user string, res Result, place, players, pos int, kind Kind) PodSeat {
	p, t := place, pos
	return PodSeat{GameID: game, UserID: user, Result: res, Place: &p, Players: players, TurnPosition: &t, WinKind: kind}
}

func TestFoldPodCountsOnlyGamesTwoSharingMembersSat(t *testing.T) {
	seats := []PodSeat{
		// g1: a and b sat; a won.
		podSeat("g1", "a", ResultWin, 1, 4, 1, KindCombat), podSeat("g1", "b", ResultLoss, 2, 4, 2, KindCombat),
		// g2: only a (b's friend c does not share): not a pod game.
		podSeat("g2", "a", ResultWin, 1, 4, 1, KindCombo),
		// g3: a, b and c's id is not in sharing; b won.
		podSeat("g3", "a", ResultLoss, 3, 4, 1, KindNoncombat), podSeat("g3", "b", ResultWin, 1, 4, 3, KindNoncombat),
		podSeat("g3", "c", ResultLoss, 2, 4, 2, KindNoncombat),
	}
	got := FoldPod([]string{"a", "b"}, seats)
	if got.Games != 2 {
		t.Fatalf("pod games %d, want 2", got.Games)
	}
	a, b := got.Members[0], got.Members[1]
	if a.PodGame.Games != 2 || a.PodGame.Wins != 1 || a.Overall.Games != 3 || a.Overall.Wins != 2 {
		t.Errorf("a: pod %+v overall %+v", a.PodGame, a.Overall)
	}
	if b.PodGame.Games != 2 || b.PodGame.Wins != 1 {
		t.Errorf("b: %+v", b.PodGame)
	}
	if len(got.WinKinds) != 2 {
		t.Errorf("win kinds %+v: the combo game was not a pod game", got.WinKinds)
	}
	if len(got.HeadToHead) != 1 || got.HeadToHead[0].AWins != 1 || got.HeadToHead[0].BWins != 1 || got.HeadToHead[0].Games != 2 {
		t.Errorf("head to head %+v", got.HeadToHead)
	}
	// Four seats at both pod games: positions 1,2 (g1) and 1,3 (g3).
	if len(got.TurnOrder) != 3 || got.TurnOrder[0].Position != 1 {
		t.Errorf("turn order %+v", got.TurnOrder)
	}
}

func TestRankPodOrdersByWinRateThenPlaceThenGames(t *testing.T) {
	place := func(v float64) *float64 { return &v }
	in := []PodRank{
		{"slow", Tally{Games: 5, Wins: 2, AvgPlace: place(2.5)}},
		{"best", Tally{Games: 6, Wins: 3, AvgPlace: place(2.0)}},
		{"tied", Tally{Games: 6, Wins: 3, AvgPlace: place(2.0)}},
		{"close", Tally{Games: 4, Wins: 4, AvgPlace: place(1)}},
		{"far", Tally{Games: 1}},
		{"better-place", Tally{Games: 5, Wins: 2, AvgPlace: place(2.0)}},
	}
	ranked, ranks, rest := RankPod(in)
	var names []string
	for _, r := range ranked {
		names = append(names, r.UserID)
	}
	want := []string{"best", "tied", "better-place", "slow"}
	if len(names) != len(want) {
		t.Fatalf("ranked %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("ranked %v, want %v", names, want)
		}
	}
	if ranks[0] != 1 || ranks[1] != 1 || ranks[2] != 3 || ranks[3] != 4 {
		t.Errorf("ranks %v, want 1 1 3 4", ranks)
	}
	if len(rest) != 2 || rest[0].UserID != "close" {
		t.Errorf("unranked %+v, want close (4 games) first", rest)
	}
}
