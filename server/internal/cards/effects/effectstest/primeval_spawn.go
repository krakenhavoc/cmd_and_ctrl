package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// primeval_spawn_test.go — Primeval Spawn (#2017): the entry clause, and the
// ten exiled cards cast against one running total mana value of 10.

const PrimevalSpawnOracle = "a11c6601-7e91-4def-bfea-bb34023ef3c6"

// spawnLeaves puts Primeval Spawn onto the battlefield, destroys it and
// settles the leaves-the-battlefield trigger, leaving the exile offered.
func SpawnLeaves(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	spawn := PushCatalogPermanent(g, me.ID, "Primeval Spawn", "Creature — Avatar", PrimevalSpawnOracle, false)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spawn); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	PassPriorityAroundTable(t, g)
	return spawn
}
