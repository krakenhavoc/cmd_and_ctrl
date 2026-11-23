package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_exile_cards_test.go — the proof cards for #1318 (a spell exiled
// from the stack takes its stack record with it; "exile target spell";
// plot) and #1320 (a zone change records what caused it).

const (
	AvenInterrupterOracle = "d31fd12f-b4dd-4bc3-ace4-703dabd0f607"
	RanarOracle           = "c73a9939-0742-4919-94d6-c3b537697f17"
)

// passUntilOnBattlefield passes priority until `id` has resolved onto
// the battlefield, leaving everything beneath it on the stack.
func PassUntilOnBattlefield(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && !g.Battlefield.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("the creature never resolved")
	}
}

// assertTableMovesOn passes priority around an empty stack and fails
// unless the step changes — the #1318 wedge, observed the way a player
// saw it.
func AssertTableMovesOn(t *testing.T, g *game.Game) {
	t.Helper()
	if len(g.StackMeta) != 0 {
		t.Fatalf("%d stack records remain with nothing on the stack", len(g.StackMeta))
	}
	start := g.Turn.Step
	for i := 0; i < 2*len(g.Seats); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		if g.Turn.Step != start {
			return
		}
	}
	t.Fatalf("the table is wedged in %q", start)
}
