package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_graveyard_search_bot_test.go — #3217. A bot at a "search your
// library and/or graveyard" prompt is enumerated both zones' cards and,
// because "and/or" lets the searcher look only in the library, fail to
// find too: a policy that prefers it may use it. Not behind
// AISEAT_GAME_TESTS: one prompt, milliseconds.

func stockGraveyard(me *game.Player, c game.Card) uuid.UUID {
	c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
	me.Graveyard.PushTop(c)
	return c.InstanceID
}

func TestHeuristicSeatAnswersALibraryAndGraveyardSearch(t *testing.T) {
	g := newSettledTable(t, 3217)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	gy := stockGraveyard(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	if !owesChoice(g, me.ID) {
		t.Fatal("two matches in two zones did not prompt")
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)

	if owesChoice(g, me.ID) {
		t.Fatal("the seat still owes the search")
	}
	if !me.Hand.Contains(lib[0]) && !me.Hand.Contains(gy) {
		t.Error("the heuristic seat found nothing although a public match was on offer")
	}
	// #4519: the zones come first, and the seat cannot see what is where,
	// so it searches everything on offer.
	if len(taken) != 2 || !strings.Contains(taken[0], "library and graveyard") {
		t.Errorf("the heuristic seat answered %q, want the zone question first with every zone", taken)
	}
}

// #4519: "search your hand and/or library" asks the zones too, and the
// heuristic seat searches both the same way.
func TestHeuristicSeatAnswersAHandAndLibrarySearch(t *testing.T) {
	g := newSettledTable(t, 4519)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, AlsoHand: true,
	})
	if !owesChoice(g, me.ID) {
		t.Fatal("a hand and/or library search did not prompt")
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)

	if owesChoice(g, me.ID) {
		t.Fatal("the seat still owes the search")
	}
	if len(taken) != 2 || !strings.Contains(taken[0], "library and hand") {
		t.Errorf("the heuristic seat answered %q, want the zone question first with every zone", taken)
	}
	if !g.Battlefield.Contains(lib[0]) {
		t.Error("the heuristic seat did not take the library match")
	}
}

// A policy may also choose one zone: the zone question offers each
// subset, and the engine takes any of them.
func TestBotMayChooseASingleZone(t *testing.T) {
	g := newSettledTable(t, 4520)
	me := g.Seats[g.Turn.ActiveSeat]
	stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	gy := stockGraveyard(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})
	pol := &graveyardOnlyPolicy{}
	driveChoices(t, g, pol, me.ID, 4)
	if pol.zoneSteps != 1 {
		t.Fatalf("the policy never saw a graveyard-only answer to the zone question (%d)", pol.zoneSteps)
	}
	if owesChoice(g, me.ID) {
		t.Fatal("the seat still owes the search")
	}
	if !me.Hand.Contains(gy) {
		t.Error("the graveyard match was not taken")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary && ev.Actor == me.ID {
			t.Error("a graveyard-only search announced a library search")
		}
	}
}

// graveyardOnlyPolicy answers the zone question with the graveyard alone
// and every later question by taking a card when it can.
type graveyardOnlyPolicy struct{ zoneSteps int }

func (p *graveyardOnlyPolicy) Name() string { return "graveyard-only" }

func (p *graveyardOnlyPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if strings.HasSuffix(m.Label, "search your graveyard only") {
			p.zoneSteps++
			return aiseat.Decision{Index: i}, nil
		}
		if strings.Contains(m.Label, ": take ") {
			return aiseat.Decision{Index: i}, nil
		}
	}
	return aiseat.Decision{Index: 0}, nil
}

func TestBotMayFailToFindWithAGraveyardMatchOnOffer(t *testing.T) {
	g := newSettledTable(t, 3218)
	me := g.Seats[g.Turn.ActiveSeat]
	stockSearchLibrary(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	stockGraveyard(me, game.Card{Name: "Command Tower", TypeLine: "Land"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Command Tower" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true, AlsoGraveyard: true,
	})

	pol := &failToFindPolicy{}
	driveChoices(t, g, pol, me.ID, 4)

	if pol.failed != 1 {
		t.Error("fail to find was not on the menu although a stated-quality search may find nothing")
	}
	if owesChoice(g, me.ID) {
		t.Error("the seat still owes the search")
	}
}
