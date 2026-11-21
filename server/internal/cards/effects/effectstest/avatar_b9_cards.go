package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AvB9Draw(t *testing.T, g *game.Game, who uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(who, n); err != nil {
			t.Fatalf("DrawNForEffect: %v", err)
		}
	})
	PassPriorityAroundTable(t, g)
}

func AvB9Sacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	PassPriorityAroundTable(t, g)
}

// avB9Live is a battlefield card with its layered characteristics read
// through a snapshot, so the layer pass has run.
func AvB9Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				_ = c.Effective()
				out, found = c, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return out
}
