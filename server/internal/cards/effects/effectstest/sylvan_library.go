package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sylvan_library_test.go — the chain, end to end.
//
// Sylvan Library is the card that named the gap in #74, and what these
// tests are really pinning is not the card: it is that a prompt can
// queue a prompt, that the second one's content depends on the first
// one's answer, and that nothing is ever asked before the player has
// answered what came before it.

const SylvanLibraryOracle = "92eed395-62ca-4293-882b-8565c40daab5"

// advanceToDrawStepOf walks the turn engine to the given seat's draw
// step, where EventBeginDrawStep fires.
func AdvanceToDrawStepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 300; i++ {
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward draw step of seat %d: %v", seat, err)
		}
	}
	t.Fatalf("never reached the draw step of seat %d", seat)
}

func ChooseCardsChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// sylvanOpening puts a Sylvan Library on seat 1's battlefield, runs the
// table to that seat's draw step, accepts the "draw two additional
// cards" offer and returns the card-set prompt that follows.
func SylvanOpening(t *testing.T, g *game.Game) (*game.Player, *game.PendingChoice) {
	t.Helper()
	owner := g.Seats[1]
	SeedLibrary(owner, "Turn", "Extra1", "Extra2", "Spare1", "Spare2", "Spare3")
	PushPermanentForTest(g, owner.ID, "Sylvan Library", SylvanLibraryOracle, "Enchantment")

	AdvanceToDrawStepOf(t, g, 1)
	// Link 1: "you may draw two additional cards" — the trigger's own
	// CR 603.5 prompt, asked as the trigger resolves.
	offer := AwaitTriggerPrompt(t, g, owner.ID)
	if offer == nil {
		t.Fatal("Sylvan Library offered no draw-step prompt")
	}
	if err := g.ResolveTriggerPrompt(offer.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	PassPriorityAroundTable(t, g)

	pick := ChooseCardsChoiceFor(g, owner.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt after the extra draws")
	}
	return owner, pick
}
