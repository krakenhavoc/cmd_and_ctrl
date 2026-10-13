package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func nezumiShortfangRow() cards.Card {
	return flipRow(nezumiShortfangOracleID, "{1}{B}", []string{"B"}, nil,
		cards.CardFace{Name: "Nezumi Shortfang", TypeLine: "Creature — Rat Rogue",
			OracleText: "{1}{B}, {T}: Target opponent discards a card. Then if that player has no cards in hand, flip this creature.",
			Power:      "1", Toughness: "1"},
		cards.CardFace{Name: "Stabwhisker the Odious", TypeLine: "Legendary Creature — Rat Shaman",
			OracleText: "At the beginning of each opponent's upkeep, that player loses 1 life for each card fewer than three in their hand.",
			Power:      "3", Toughness: "3"})
}

// flipFillHand replaces a player's hand with n plain cards and returns their
// ids.
func flipFillHand(g *game.Game, p *game.Player, n int) []uuid.UUID {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, handCardForTest(p, "Filler", "Sorcery", ""))
	}
	return ids
}

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
			g, me, opp := flipMain(t)
			id := flipEntered(t, g, nezumiShortfangRow(), me)
			hand := flipFillHand(g, opp, tc.hand)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if discardChoiceFor(g, opp.ID) != nil {
				// The opponent picks the card (the last one, to prove it is
				// their choice and not the first).
				answerDiscard(t, g, opp.ID, hand[len(hand)-1])
			}
			passPriorityAroundTable(t, g)
			if want := max(tc.hand-1, 0); opp.Hand.Size() != want {
				t.Fatalf("the opponent holds %d cards, want %d", opp.Hand.Size(), want)
			}
			if tc.hand > 0 && !opp.Graveyard.Contains(hand[len(hand)-1]) {
				t.Error("the card the opponent chose is not in their graveyard")
			}
			if tc.wantFlip {
				wantFlipped(t, g, id, "Stabwhisker the Odious", []string{"B"}, 2)
			} else {
				wantUnflipped(t, g, id, "Nezumi Shortfang")
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
		g, me, opp := flipMain(t)
		id := flipEntered(t, g, nezumiShortfangRow(), me)
		g.WithWriteLock(func() {
			if err := g.FlipPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		wantFlipped(t, g, id, "Stabwhisker the Odious", []string{"B"}, 2)
		flipFillHand(g, opp, tc.hand)
		flipFillHand(g, me, 0)
		life, mine := opp.Life, me.Life
		advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
		passPriorityAroundTable(t, g)
		if got := life - opp.Life; got != tc.wantLoss {
			t.Errorf("hand %d: the opponent lost %d, want %d", tc.hand, got, tc.wantLoss)
		}
		if me.Life != mine {
			t.Errorf("hand %d: Stabwhisker's controller lost life (%d → %d)", tc.hand, mine, me.Life)
		}
	}
}
