package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b216_helpers.go — shared shapes for GAP batch 216 (#5948).

// GapB216DragonsTotalManaValue is "the total mana value of Dragons you
// control" (Earthquake Dragon's cost reduction), read as the spell is cast: the
// caster's battlefield Dragons, each at its mana value (a token has none, so 0).
func GapB216DragonsTotalManaValue(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	total := 0
	for i := range q.Game.Battlefield.Cards {
		c := &q.Game.Battlefield.Cards[i]
		if c.Controller == q.Controller && c.HasSubtypeWord("Dragon") {
			total += c.ManaValue()
		}
	}
	return total
}
