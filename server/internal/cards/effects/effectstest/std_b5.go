package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b5_test.go — Standard set batches 5 and 6 (Edge of Eternities
// leftovers, umbrella #4452).

func RequireFullNoCaveats(t *testing.T, oracle, name string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
	return spec
}

func CounterCountOf(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	n, found := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n, found = c.Counters[kind], true
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return n
}
