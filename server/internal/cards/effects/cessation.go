package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cessation — Enchantment — Aura {2}{W} (Urza's Legacy, #5623):
//
//	"Enchant creature
//	 Enchanted creature can't attack.
//	 When this Aura is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// Despondency's return on a can't-attack Aura, through the shared
// ulgB1AuraReturnsToHand trigger. The restriction is checked as attackers
// are declared; the creature can still block.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8b3a35e-0172-4f7b-b50f-571e59d7ca5f",
		Name:         "Cessation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttack),
		},
		Triggered: []game.TriggeredAbility{
			ulgB1AuraReturnsToHand("Cessation"),
		},
	})
}
