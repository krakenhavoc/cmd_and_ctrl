package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Homicidal Seclusion — {4}{B} Enchantment: "As long as you control exactly
// one creature, that creature gets +3/+1 and has lifelink."
//
// Both halves are layered statics that re-read the count every pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bc51b45-1f96-4277-8b80-bb0a528c6ac2",
		Name:         "Homicidal Seclusion",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.ControlsExactlyOneCreatureAndIsIt, 3, 1),
			effects.B16GrantKeywords(effects.ControlsExactlyOneCreatureAndIsIt, "lifelink"),
		},
	})
}
