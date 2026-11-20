package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmtB1Count counts the battlefield permanents `controller` controls with
// the given name.
func TmtB1Count(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() { n = B27CountNamed(g, controller, name) })
	return n
}

// tmtB1Card reads a battlefield card, failing the test when it is gone.
func TmtB1Card(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

// tmtB1Named returns the ids of the battlefield permanents with a name.
func TmtB1Named(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}
