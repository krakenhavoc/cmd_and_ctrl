package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// onlyStackItemOf is the single item on the stack, or a failure.
func OnlyStackItemOf(t *testing.T, g *game.Game) *game.StackItem {
	t.Helper()
	if len(g.StackMeta) != 1 {
		t.Fatalf("stack holds %d items, want 1", len(g.StackMeta))
	}
	for _, it := range g.StackMeta {
		return it
	}
	return nil
}

// pfAddCounters puts n counters of `kind` on a battlefield card.
func PfAddCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID == id {
				if c.Counters == nil {
					c.Counters = map[string]int{}
				}
				c.Counters[kind] += n
			}
		}
	})
	g.BumpLayerVersionForTest()
}

func PfController(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			out = c.Controller
		}
	})
	return out
}
