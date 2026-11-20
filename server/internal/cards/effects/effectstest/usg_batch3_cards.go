package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usgB3Colored seeds a creature of the given colours, not summoning sick.
func UsgB3Colored(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Colors: colors, Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// usgB3Destroy destroys a permanent and lets what it triggered resolve.
func UsgB3Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	PassPriorityAroundTable(t, g)
}
