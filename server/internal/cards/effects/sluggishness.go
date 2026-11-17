package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sluggishness — Enchantment — Aura {1}{R} (Urza's Legacy, #5623):
//
//	"Enchant creature
//	 Enchanted creature can't block.
//	 When this Aura is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Despondency's return on a can't-block Aura, through the shared
// ulgB1AuraReturnsToHand trigger. The restriction is checked as blockers are
// declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39eb4062-68a0-4f64-a9cc-79be87fd8173",
		Name:         "Sluggishness",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			ulgB1AuraReturnsToHand("Sluggishness"),
		},
	})
}
