package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dropped_card_set_pick_test.go — #1225 at the cards.
//
// dropped_option_pick_test.go beside this pins #1006 with the same
// card: a dropped OPTION PICK runs its continuation, so Torment of
// Hailfire asks the opponents after the one who left. This is the
// sacrifice BRANCH of that same question, one prompt further in.
//
// effects.SacrificeChoice queues a choose_cards over the victim's
// nonland permanents with a floor of one, and its `Then` is the rest of
// the card — the sacrifice, then the next repetition. It is no run's
// leg and carries no option frame, so before #1225 the withdrawal ran
// nothing at all and the card stopped at the victim whose board
// emptied, which is the exact sentence #1006 wrote about the other
// half of this card.

// TestTormentOfHailfireRunsOnPastAWithdrawnSacrificePrompt.
func TestTormentOfHailfireRunsOnPastAWithdrawnSacrificePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	bear := b12Creature(g, first.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := second.Life

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	// The first victim takes the sacrifice branch, which chains a
	// mid-card card pick over their own board.
	answerOptionPick(t, g, first.ID, 1)
	sac := latestChooseCardsFor(g, first.ID)
	if sac == nil {
		t.Fatalf("the sacrifice branch asks which permanent: %+v", g.PendingChoices)
	}
	if latestOptionPickFor(g, second.ID) != nil {
		t.Fatal("setup: the next victim must not be asked until this one has finished")
	}

	// Their last nonland permanent leaves — somebody's removal in
	// response to nothing, an admin move, a state-based death — while
	// the question is open. The prompt has no legal answer left and is
	// withdrawn (#1045).
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bear); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})

	if latestChooseCardsFor(g, first.ID) != nil {
		t.Fatal("the pick survives a board with none of its candidates on it")
	}
	if first.Life != life {
		t.Errorf("a withdrawn sacrifice does not fall back to the life branch: life is %d", first.Life)
	}

	pick := latestOptionPickFor(g, second.ID)
	if pick == nil {
		t.Fatalf("the run stopped at the victim whose board emptied — the rest of the card "+
			"went with the withdrawn pick (#1225). Open prompts: %+v", g.PendingChoices)
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

// TestTormentOfHailfireRunsOnPastASacrificePromptWhoseVictimLeaves is
// the other door into the same drop action — the departure sweep —
// reached at the sacrifice prompt rather than at the option pick.
func TestTormentOfHailfireRunsOnPastASacrificePromptWhoseVictimLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	b12Creature(g, first.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := second.Life

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)

	answerOptionPick(t, g, first.ID, 1)
	if latestChooseCardsFor(g, first.ID) == nil {
		t.Fatalf("the sacrifice branch asks which permanent: %+v", g.PendingChoices)
	}

	// CR 800.4a takes their board out of the game with them, so there
	// is nothing for anybody else to choose; the sorcery is a
	// survivor's and still owes the seats after them their question.
	if err := g.Concede(first.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if latestChooseCardsFor(g, first.ID) != nil {
		t.Fatal("a pick over the departed seat's own board is dropped, not reassigned")
	}
	if latestOptionPickFor(g, second.ID) == nil {
		t.Fatalf("the run stopped at the seat that left mid-sacrifice: %+v", g.PendingChoices)
	}

	answerOptionPick(t, g, second.ID, 0)
	if second.Life != life-3 {
		t.Errorf("the second opponent's life %d → %d, want -3", life, second.Life)
	}
	drainRemainingTormentPrompts(t, g)
}
