package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witness Protection — Enchantment — Aura {U} (Foundations, #4465):
//
//	"Enchant creature
//	 Enchanted creature loses all abilities and is a green and white
//	 Citizen creature with base power and toughness 1/1 named
//	 Legitimate Businessperson. (It loses all other colors, card types,
//	 creature types, and names.)"
//
// One effect across five layers (CR 613.6), each in the layer the rules
// put it, as Honest Work's is:
//
//   - layer 3: named Legitimate Businessperson (CR 612.8);
//   - layer 4: its card types become creature alone and its subtypes
//     Citizen alone, so an Equipment or Vehicle subtype goes too (the
//     ruling); supertypes stay (CR 205.1a);
//   - layer 5: green and white, every other colour gone;
//   - layer 6: loses all abilities;
//   - layer 7b: base power and toughness 1/1, so counters and pumps
//     still apply on top and an earlier set is overwritten.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5f5e71e-83a7-468d-829a-6392b764c76e",
		Name:         "Witness Protection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.AvB4SetAttachedName("Legitimate Businessperson"),
			effects.SetAttachedTypes([]string{"Creature"}, []string{"Citizen"}),
			effects.SetAttachedColors("G", "W"),
			effects.LoseAllAbilities(),
			effects.SetAttachedBasePT(1, 1),
		},
	})
}
