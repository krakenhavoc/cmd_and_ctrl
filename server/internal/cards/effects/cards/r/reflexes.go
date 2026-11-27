package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reflexes — Enchantment — Aura {R} (Eighth Edition, S82 older sets batch 617,
// #5607):
//
//	"Enchant creature
//	 Enchanted creature has first strike."
//
// Flight's grant with the first strike keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc87b0a5-3d9d-44eb-b415-b022acd63cf1",
		Name:         "Reflexes",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("first strike")},
	})
}
