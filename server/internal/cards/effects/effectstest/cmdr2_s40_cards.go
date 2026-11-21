package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func S40SuspendAction(t *testing.T, oracle string, n int, cost string) {
	t.Helper()
	spec, _ := effects.Lookup(oracle)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	sa := spec.SpecialActions[0]
	if sa.Kind != game.SpecialActionSuspend || sa.Counters != n || sa.Cost != cost {
		t.Errorf("special action %+v, want suspend %d—%s", sa, n, cost)
	}
}
