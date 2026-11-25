package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusory Wrappings — Enchantment — Aura {2}{U} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 0/2."
//
// Layer 7b: it sets the base, so counters and other pumps still apply on top
// (CR 613.4b, 613.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bda363b6-4be1-4415-8b3e-92882df1f630",
		Name:         "Illusory Wrappings",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.SetAttachedBasePT(0, 2)},
	})
}
