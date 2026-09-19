package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// option_pick_test.go — #568's enumerator half, and the reason it
// exists: `EnumerateFor` returns ONLY a choice's answers while a seat
// owes one, so a kind the enumerator does not know is a bot seat with
// an empty move list holding a live table (#544, #499).
//
// The extra thing this kind has to get right is WHOSE seat it is. An
// option pick is addressed to an opponent, and a prompt enumerated for
// the wrong seat would offer the controller an answer the engine
// refuses and the chooser nothing at all.

// TestOptionPickEnumeratesEveryBranchForItsChooser — and dispatchAll
// proves the engine accepts every one of them.
func TestOptionPickEnumeratesEveryBranchForItsChooser(t *testing.T) {
	g := newTable(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  opp.ID,
			Question: "Torment of Hailfire",
			Options: []game.ChoiceOption{
				{Label: "Lose 3 life", LifeCost: 3},
				{Label: "Sacrifice a nonland permanent"},
				{Label: "Discard a card"},
			},
			Then: func(*game.Game, int) error { return nil },
		})
	})

	moves := legal.EnumerateFor(g, opp.ID)
	if len(moves) != 3 {
		t.Fatalf("enumerated %d answers, want 3: %v", len(moves), labels(moves))
	}
	for _, want := range []string{
		"Torment of Hailfire: Lose 3 life",
		"Torment of Hailfire: Sacrifice a nonland permanent",
		"Torment of Hailfire: Discard a card",
	} {
		if !hasLabel(moves, want) {
			t.Errorf("missing %q: %v", want, labels(moves))
		}
	}
	dispatchAll(t, g, opp.ID, moves)
}

// TestOptionPickPricesTheLifeBranch — #547. A policy holding only the
// wire payload must be able to tell "lose 3 life" from "discard a
// card", or a bot at 3 life answers with the life and dies.
func TestOptionPickPricesTheLifeBranch(t *testing.T) {
	g := newTable(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  opp.ID,
			Question: "Torment of Hailfire",
			Options: []game.ChoiceOption{
				{Label: "Discard a card"},
				{Label: "Lose 3 life", LifeCost: 3},
			},
			Then: func(*game.Game, int) error { return nil },
		})
	})
	moves := legal.EnumerateFor(g, opp.ID)
	for _, m := range moves {
		switch m.Label {
		case "Torment of Hailfire: Lose 3 life":
			if m.Cost == nil || m.Cost.Life != 3 {
				t.Errorf("the life branch is priced: %+v", m.Cost)
			}
		case "Torment of Hailfire: Discard a card":
			// moveCost returns nil when nothing is charged, so a free
			// branch stays byte-identical on the wire.
			if m.Cost != nil {
				t.Errorf("the discard branch costs no life: %+v", m.Cost)
			}
			if !m.AlwaysLegal {
				t.Error("the FIRST option is the always-legal way out of the kind")
			}
		}
	}
}

// TestOptionPickOffersNothingToAnyoneElse — the prompt belongs to one
// seat. The controller of the effect that asked it is not offered its
// answers, and is not blocked out of their own turn either.
func TestOptionPickOffersNothingToAnyoneElse(t *testing.T) {
	g := newTable(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  opp.ID,
			Question: "Torment of Hailfire",
			Options:  []game.ChoiceOption{{Label: "Lose 3 life", LifeCost: 3}},
			Then:     func(*game.Game, int) error { return nil },
		})
	})
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeResolveChoice {
			t.Errorf("another seat's prompt was offered to the controller: %q", m.Label)
		}
	}
	// And the chooser gets its answers and nothing else, because the
	// kind blocks the table.
	moves := legal.EnumerateFor(g, opp.ID)
	if len(moves) == 0 {
		t.Fatal("the seat owing the pick was offered nothing at all")
	}
	for _, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("a seat owing a blocking choice was offered %q as well", m.Label)
		}
	}
}
