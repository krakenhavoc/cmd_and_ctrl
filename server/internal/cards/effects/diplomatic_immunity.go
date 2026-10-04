package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diplomatic Immunity — Enchantment — Aura {1}{U}:
//
//	"Enchant creature
//	 Shroud (A permanent with shroud can't be the target of spells or
//	 abilities.)
//	 Enchanted creature has shroud."
//
// Two shrouds: the Aura's own (a printed keyword, so nothing can
// target the Aura to remove it) and the enchanted creature's (a layer-6
// grant through GrantToAttached, Whispersilk Cloak's shape). Shroud
// (CR 702.18a) stops every player's spells and abilities, the Aura's
// controller's included.
//
// The Aura spell itself can still be targeted while it is on the stack:
// shroud applies only once it is on the battlefield (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1a3153c-0200-4ff2-b7d5-48d23920bb3c",
		Name:            "Diplomatic Immunity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("shroud"),
		},
	})
}
