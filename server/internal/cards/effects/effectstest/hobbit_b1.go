package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4454, The Hobbit (HOB) cards, half B, batch 1. Every test drives the
// registered catalog row through the engine.

func HobOracle(t *testing.T, name string) string {
	t.Helper()
	for _, s := range effects.All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	t.Fatalf("%s is not registered", name)
	return ""
}

// hobPush puts the named catalog creature onto the battlefield, ready to
// attack.
func HobPush(t *testing.T, g *game.Game, who uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	t.Helper()
	return B12Push(g, who, name, typeLine, HobOracle(t, name), power, toughness)
}

func HobArmyCounters(t *testing.T, g *game.Game, who uuid.UUID) int {
	t.Helper()
	return ArmyOf(t, g, who).Counters[game.CounterPlusOne]
}

func HobDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	PassPriorityAroundTable(t, g)
}
