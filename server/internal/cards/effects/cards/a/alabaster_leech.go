package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alabaster Leech — Creature — Leech {W}, 1/3 (Invasion, S82 older sets batch 664, #5616):
//
//	"White spells you cast cost {W} more to cast."
//
// Jade Leech's tax in white: the controller's white spells cost {W} more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a8b032e-90b7-4782-b854-b575f316d641",
		Name:         "Alabaster Leech",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.OldB664CostsMoreMana("{W}", "White spells you cast cost {W} more to cast.", effects.YourSpell(), effects.ColoredSpell("W")),
		},
	})
}
