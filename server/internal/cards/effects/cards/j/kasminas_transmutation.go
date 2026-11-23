package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kasmina's Transmutation — Enchantment — Aura {1}{U} (Ravnica Remastered):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and has base power and
//	 toughness 1/1."
//
// Amphibian Downpour's two layers: the ability loss (layer 6) and the
// base-P/T set (layer 7b), in timestamp order with the Aura's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9be16d1-5e53-4b85-8921-c6fe9473f923",
		Name:         "Kasmina's Transmutation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
	})
}
