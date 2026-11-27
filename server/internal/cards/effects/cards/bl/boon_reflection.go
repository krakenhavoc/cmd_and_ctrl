package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boon Reflection — Enchantment {4}{W} (Shadowmoor, S82 OLD b529, #5591):
//
//	"If you would gain life, you gain twice that much life instead."
//
// Rhox Faithmender's replacement on its own: a CR 614 replacement on the
// life-change event for the controller's gains (YouGainTwiceThatMuchLife),
// so it doubles a catalog GainLife, a drain's gain half and lifelink alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b86d3142-568a-4eea-b1ae-5fcfc1453533",
		Name:         "Boon Reflection",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.YouGainTwiceThatMuchLife("Boon Reflection: gain twice that much life"),
		},
	})
}
