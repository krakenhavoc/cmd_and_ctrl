package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sleeper's Guile — Enchantment — Aura {2}{B} (Urza's Legacy, #5623):
//
//	"Enchant creature
//	 Enchanted creature has fear. (It can't be blocked except by
//	 artifact creatures and/or black creatures.)
//	 When this Aura is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Rancor's return on a fear Aura, through the shared ulgB1AuraReturnsToHand
// trigger. Fear is a printed keyword the block check reads (CR 702.36b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eaaa3e41-dbf6-43ae-b65e-e1502ba4fa0b",
		Name:         "Sleeper's Guile",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached("fear"),
		},
		Triggered: []game.TriggeredAbility{
			ulgB1AuraReturnsToHand("Sleeper's Guile"),
		},
	})
}
