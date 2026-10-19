package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dont_search_bot_test.go — #4134. A "you may search" prompt offers a
// move beyond the card picks and "fail to find": not searching at all.
// The enumerator must offer it (a bot picks from the closed list), the
// heuristic must still prefer taking the card, and a policy that does
// pick it must be accepted, play on, and leave the library unshuffled.

// dontSearchPolicy picks the "don't search" move when it is on the
// menu as the always-legal answer, and otherwise the first move.
type dontSearchPolicy struct{ declined, other int }

func (p *dontSearchPolicy) Name() string { return "dont-search" }

func (p *dontSearchPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if m.AlwaysLegal && strings.Contains(m.Label, "don't search") {
			p.declined++
			return aiseat.Decision{Index: i}, nil
		}
	}
	p.other++
	return aiseat.Decision{Index: 0}, nil
}

func optionalLandSearch(me *game.Player) game.SearchLibrarySpec {
	return game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.IsLand() },
		Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, Optional: true,
	}
}

func TestBotMayDeclineAYouMaySearchAndTheTableCarriesOn(t *testing.T) {
	g := newSettledTable(t, 4134)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stockSearchLibrary(me,
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	before := make([]string, 0, len(ids))
	for _, c := range me.Library.Cards {
		before = append(before, c.InstanceID.String())
	}
	searchForBot(t, g, optionalLandSearch(me))
	if !owesChoice(g, me.ID) {
		t.Fatal("no prompt")
	}

	pol := &dontSearchPolicy{}
	taken := driveChoices(t, g, pol, me.ID, 4)

	if pol.declined != 1 {
		t.Fatalf("the policy never saw \"don't search\" as the always-legal answer (took %v)", taken)
	}
	if owesChoice(g, me.ID) {
		t.Error("the seat still owes a choice after declining")
	}
	if g.Battlefield.Contains(ids[1]) || !me.Library.Contains(ids[1]) {
		t.Error("declining still fetched the land")
	}
	for i, c := range me.Library.Cards {
		if c.InstanceID.String() != before[i] {
			t.Fatal("declining shuffled the library")
		}
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == me.ID {
			t.Error("declining announced a search")
		}
	}
}

func TestHeuristicSeatSearchesWhenItMay(t *testing.T) {
	g := newSettledTable(t, 4135)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stockSearchLibrary(me,
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	searchForBot(t, g, optionalLandSearch(me))
	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if !g.Battlefield.Contains(ids[1]) {
		t.Errorf("the heuristic seat declined a land it was free to fetch (took %v)", taken)
	}
	// Two answers: "search" (the question comes first), then the take.
	if len(taken) != 2 || !strings.Contains(taken[0], ": search") || !strings.Contains(taken[1], "Forest") {
		t.Errorf("answers = %v, want the search, then taking the Forest", taken)
	}
}
