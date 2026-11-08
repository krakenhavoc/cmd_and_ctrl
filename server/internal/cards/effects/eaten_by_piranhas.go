package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eaten by Piranhas — Enchantment — Aura {1}{U} (The Lost Caverns of
// Ixalan, #4470):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature loses all abilities and is a black Skeleton
//	 creature with base power and toughness 1/1. (It loses all other
//	 colors, card types, and creature types.)"
//
// Kenrith's Transformation's static without the cantrip, one effect
// across four layers (CR 613.6):
//
//   - layer 4: "is a Skeleton creature", replacing its card types and
//     creature types (SetAttachedTypes), so it stays a creature and the
//     Equipment and Auras on it stay;
//   - layer 5: black, replacing its colours;
//   - layer 6: loses all abilities;
//   - layer 7b: base 1/1, so counters and pumps still apply on top.
//
// "Skeleton" and "black" are words of the Aura's text, read through it
// (CR 612.2). Flash is its printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c37eae65-1afe-4242-b677-24dd40e6f401",
		Name:            "Eaten by Piranhas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			SetAttachedTypes([]string{"Creature"}, []string{"Skeleton"}),
			SetAttachedColors("B"),
			LoseAllAbilities(),
			SetAttachedBasePT(1, 1),
		},
	})
}
