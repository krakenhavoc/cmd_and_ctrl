package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PushDevourFodder(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// castDevourer casts the creature and passes priority until its entry
// asks (or it lands).
func CastDevourer(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, name, "Creature", oracle, nil)
	PassPriorityAroundTable(t, g)
	return id
}

func DevourPlusOnes(g *game.Game, id uuid.UUID) int {
	c, ok := BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}
