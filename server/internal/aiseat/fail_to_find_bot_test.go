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

// fail_to_find_bot_test.go — #3053. Since a search for a stated quality
// may find nothing even against one match, "find nothing" is on every
// tutor prompt's menu, and a bot that treated it as free would leave the
// card in the library. The heuristic must still take the best card; and
// a policy that does answer "find nothing" must be accepted and play on.
//
// Not behind AISEAT_GAME_TESTS: one prompt, milliseconds.

// failToFindPolicy answers a search with the always-legal move when it
// is on the menu and otherwise the first move, recording which it used.
type failToFindPolicy struct{ failed, other int }

func (p *failToFindPolicy) Name() string { return "fail-to-find" }

func (p *failToFindPolicy) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	for i, m := range in.Moves {
		if m.AlwaysLegal && strings.Contains(m.Label, "fail to find") {
			p.failed++
			return aiseat.Decision{Index: i}, nil
		}
	}
	p.other++
	return aiseat.Decision{Index: 0}, nil
}

func stockSearchLibrary(me *game.Player, cards ...game.Card) []uuid.UUID {
	me.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func searchForBot(t *testing.T, g *game.Game, spec game.SearchLibrarySpec) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SearchLibraryThenForEffect(spec); err != nil {
			t.Fatalf("SearchLibraryThenForEffect: %v", err)
		}
	})
}

func TestHeuristicSeatTakesTheLoneMatchOfAStatedQualitySearch(t *testing.T) {
	g := newSettledTable(t, 3053)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stockSearchLibrary(me,
		game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"},
	)
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.IsLand() },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true,
	})
	if !owesChoice(g, me.ID) {
		t.Fatal("a one-match search for a stated quality did not prompt")
	}

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)

	if !me.Hand.Contains(ids[1]) {
		t.Errorf("the heuristic seat declined the only land (%v): it should take the card unless it has a reason not to", taken)
	}
}

func TestHeuristicSeatTakesAWorthlessCardRatherThanFindNothing(t *testing.T) {
	g := newSettledTable(t, 3054)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stockSearchLibrary(me, game.Card{Name: "Blank", TypeLine: "Artifact"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.Name == "Blank" },
		Dest: game.ZoneHand, Limit: 1, Shuffle: true,
	})
	driveChoices(t, g, heuristic.New(), me.ID, 4)
	if !me.Hand.Contains(ids[0]) {
		t.Error("a card valued at nothing was declined for no reason")
	}
}

func TestBotMayFindNothingAndTheTableCarriesOn(t *testing.T) {
	g := newSettledTable(t, 3055)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := stockSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Pred: func(c game.Card) bool { return c.IsLand() },
		Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true,
	})

	pol := &failToFindPolicy{}
	taken := driveChoices(t, g, pol, me.ID, 4)

	if pol.failed != 1 {
		t.Fatalf("the policy never saw fail to find on offer (took %v)", taken)
	}
	if owesChoice(g, me.ID) {
		t.Error("the seat still owes a choice after failing to find")
	}
	if !me.Library.Contains(ids[0]) || g.Battlefield.Contains(ids[0]) {
		t.Error("failing to find still took the land")
	}
}

// An unconditional search is not failable: the menu has no such move,
// and a policy preferring it falls through to a take.
func TestBotCannotFailAnUnconditionalSearch(t *testing.T) {
	g := newSettledTable(t, 3056)
	me := g.Seats[g.Turn.ActiveSeat]
	stockSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Creature — Bear"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
	)
	handBefore := me.Hand.Size()
	searchForBot(t, g, game.SearchLibrarySpec{
		Player: me.ID, Dest: game.ZoneHand, Limit: 1, Shuffle: true,
	})

	pol := &failToFindPolicy{}
	driveChoices(t, g, pol, me.ID, 4)

	if pol.failed != 0 {
		t.Error("an unconditional search offered fail to find")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d → %d, want one card found", handBefore, me.Hand.Size())
	}
}
