package testsms

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nezumi Shortfang makes an opponent discard a card of their choice,
// then flips if that player's hand is empty; an already empty hand
// flips it too.
func TestNezumiShortfangFlipsWhenTheHandIsEmpty(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hand     int
		wantFlip bool
	}{{"one card", 1, true}, {"two cards", 2, false}, {"already empty", 0, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := effectstest.FlipMain(t)
			id := effectstest.FlipEntered(t, g, effectstest.NezumiShortfangRow(), me)
			hand := effectstest.FlipFillHand(g, opp, tc.hand)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
				// The opponent picks the card (the last one, to prove it is
				// their choice and not the first).
				effectstest.AnswerDiscard(t, g, opp.ID, hand[len(hand)-1])
			}
			effectstest.PassPriorityAroundTable(t, g)
			if want := max(tc.hand-1, 0); opp.Hand.Size() != want {
				t.Fatalf("the opponent holds %d cards, want %d", opp.Hand.Size(), want)
			}
			if tc.hand > 0 && !opp.Graveyard.Contains(hand[len(hand)-1]) {
				t.Error("the card the opponent chose is not in their graveyard")
			}
			if tc.wantFlip {
				effectstest.WantFlipped(t, g, id, "Stabwhisker the Odious", []string{"B"}, 2)
			} else {
				effectstest.WantUnflipped(t, g, id, "Nezumi Shortfang")
			}
		})
	}
}

// Stabwhisker drains each opponent at their upkeep for every card fewer
// than three in their hand, and does not touch its own controller.
func TestStabwhiskerDrainsOpponentsWithSmallHands(t *testing.T) {
	for _, tc := range []struct {
		hand     int // at the upkeep, after the turn's draw (none before upkeep)
		wantLoss int
	}{{0, 3}, {1, 2}, {2, 1}, {3, 0}, {5, 0}} {
		g, me, opp := effectstest.FlipMain(t)
		id := effectstest.FlipEntered(t, g, effectstest.NezumiShortfangRow(), me)
		g.WithWriteLock(func() {
			if err := g.FlipPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		effectstest.WantFlipped(t, g, id, "Stabwhisker the Odious", []string{"B"}, 2)
		effectstest.FlipFillHand(g, opp, tc.hand)
		effectstest.FlipFillHand(g, me, 0)
		life, mine := opp.Life, me.Life
		effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		effectstest.PassPriorityAroundTable(t, g)
		if got := life - opp.Life; got != tc.wantLoss {
			t.Errorf("hand %d: the opponent lost %d, want %d", tc.hand, got, tc.wantLoss)
		}
		if me.Life != mine {
			t.Errorf("hand %d: Stabwhisker's controller lost life (%d → %d)", tc.hand, mine, me.Life)
		}
	}
}
