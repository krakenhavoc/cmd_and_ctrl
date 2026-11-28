package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entangler — Enchantment — Aura {2}{W}{W} (Prophecy, S82 OLD b669, #5617):
//
//	"Enchant creature
//	 Enchanted creature can block any number of creatures."
//
// Wall of Glare's CanBlockAnyNumber on the enchanted creature (layer 6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "285cf707-df0c-4eb9-8f60-5d052ee8eb8c",
		Name:         "Entangler",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.CanBlockAnyNumber(effects.AttachedToSource)},
	})
}
