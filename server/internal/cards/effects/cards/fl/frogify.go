package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frogify — Enchantment — Aura {1}{U} (Throne of Eldraine, S82 old sets):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and is a blue Frog creature with
//	 base power and toughness 1/1. (It loses all other card types and creature
//	 types.)"
//
// Kenrith's Transformation's layers (CR 613.6): layer 4 makes it a Frog
// creature alone, layer 5 blue alone, layer 6 removes every ability, layer 7b
// sets the base power and toughness so counters and pumps still apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32249228-e300-4865-b977-5e1f285d02f2",
		Name:         "Frogify",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.SetAttachedTypes([]string{"Creature"}, []string{"Frog"}),
			effects.SetAttachedColors("U"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
	})
}
