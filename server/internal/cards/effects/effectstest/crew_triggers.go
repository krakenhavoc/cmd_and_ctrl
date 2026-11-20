package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolveTopOfStack passes priority around the table until exactly one
// item has left the stack.
func ResolveTopOfStack(t *testing.T, g *game.Game) {
	t.Helper()
	before := len(g.StackMeta)
	for i := 0; i < 16 && len(g.StackMeta) >= before; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(g.StackMeta) >= before {
		t.Fatalf("nothing resolved: %d items before, %d after", before, len(g.StackMeta))
	}
}

func CrewEventsThisTurn(g *game.Game, kind game.EventKind) (out []game.Event) {
	g.WithWriteLock(func() {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == kind {
				out = append(out, ev)
			}
		}
	})
	return out
}
