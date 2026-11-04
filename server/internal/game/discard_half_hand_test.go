package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_half_hand_test.go — DiscardPrompt.HalfHand, "discards half the cards
// in their hand, round down" (Pox Plague, Secrets of Strixhaven, #4456): the
// count is the discarding player's own hand size halved as that player is
// asked, N is ignored, and a hand of one card is not asked at all.
func TestDiscardPromptHalfHandRoundsDownPerPlayer(t *testing.T) {
	for _, tc := range []struct {
		hand, owed int
	}{{7, 3}, {6, 3}, {2, 1}, {1, 0}, {0, 0}} {
		g := newActiveGame(t)
		victim := g.Seats[1]
		if tc.hand > victim.Hand.Size() {
			t.Fatalf("opening hand %d is smaller than the %d this case needs", victim.Hand.Size(), tc.hand)
		}
		victim.Hand.Cards = victim.Hand.Cards[:tc.hand]

		id := queueEffectDiscard(g, DiscardPrompt{Player: victim.ID, N: 99, HalfHand: true})
		c := discardPromptFor(g, victim.ID)
		if tc.owed == 0 {
			if id != uuid.Nil || c != nil {
				t.Errorf("hand %d: a prompt was queued, want none (half rounds down to 0)", tc.hand)
			}
			continue
		}
		if c == nil {
			t.Fatalf("hand %d: no discard prompt queued", tc.hand)
		}
		if c.ChooseMin != tc.owed || c.ChooseMax != tc.owed {
			t.Errorf("hand %d: bounds [%d,%d], want [%d,%d]", tc.hand, c.ChooseMin, c.ChooseMax, tc.owed, tc.owed)
		}
	}
}
