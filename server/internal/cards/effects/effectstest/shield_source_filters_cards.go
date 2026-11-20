package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sfPush puts a permanent with the given type line, power and colours on
// the battlefield for `owner`, toughness 5 so combat damage marks rather
// than kills.
func SfPush(g *game.Game, owner uuid.UUID, name, typeLine string, power int, colors []string, keywords ...string) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: 5,
		Colors: colors, Keywords: keywords})
}

// sfCombat walks to the combat damage step and back out of its priority
// boundary, returning how much life each player lost.
func SfCombat(t *testing.T, g *game.Game) map[uuid.UUID]int {
	t.Helper()
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	Pr7bDamageStep(t, g)
	PassPriorityAroundTable(t, g)
	lost := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lost[p.ID] = before[p.ID] - p.Life
	}
	return lost
}
