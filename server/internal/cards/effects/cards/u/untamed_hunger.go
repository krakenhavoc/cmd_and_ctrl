package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Untamed Hunger — Enchantment — Aura {2}{B} (Foundations, #4465):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has menace."
//
// Furious Strength's shape with the printed numbers: a layer 7c pump
// and a layer 6 menace grant to the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "550abd9b-9aae-4812-9bc0-7f131d6dcde8",
		Name:         "Untamed Hunger",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1), effects.GrantToAttached("menace")},
	})
}
