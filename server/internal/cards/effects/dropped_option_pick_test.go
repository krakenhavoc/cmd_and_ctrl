package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dropped_option_pick_test.go — #1006 at the cards.
//
// The engine rule is that a dropped option pick runs its continuation
// with the no-choice outcome. Torment of Hailfire is the card that
// makes the difference visible, because its continuation is not a
// tidy-up: it is the NEXT question. The run is a chain — X repetitions
// × the opponents, each asked by the answer to the one before it — so
// a prompt that ends without running its continuation ends the card,
// and the opponents after the one who left are never asked at all.

// TestTormentOfHailfireRunsOnPastAVictimWhoConcedes.
func TestTormentOfHailfireRunsOnPastAVictimWhoConcedes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	life := second.Life

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	if latestOptionPickFor(g, first.ID) == nil {
		t.Fatalf("the first opponent was not asked: %+v", g.PendingChoices)
	}
	if latestOptionPickFor(g, second.ID) != nil {
		t.Fatal("setup: the questions are supposed to be asked one at a time")
	}

	// The victim leaves with the question in front of them. CR 800.4a
	// takes their material out of the game, so there is no answer for
	// anybody else to give — but the rest of the card is still
	// resolving, and the next opponent is owed their question.
	if err := g.Concede(first.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	pick := latestOptionPickFor(g, second.ID)
	if pick == nil {
		t.Fatalf("the run stopped at the seat that left — the rest of the card was dropped "+
			"with the question (#1006). Open prompts: %+v", g.PendingChoices)
	}

	// And it is a real question that finishes the card.
	answerOptionPick(t, g, second.ID, 0)
	if second.Life != life-3 {
		t.Errorf("the second opponent's life %d → %d, want -3", life, second.Life)
	}
	drainRemainingTormentPrompts(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			t.Errorf("an option pick is still open for %s after the run ended", c.Chooser)
		}
	}
}
