package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Containment Membrane — Enchantment — Aura {2}{U} (Oath of the Gatewatch,
// #5545):
//
//	"Surge {U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Surge is the shared gated alternative cost (CR 702.117a); the lock is Bind
// the Monster's EnchantedDoesntUntap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "e3bf1a95-4363-4915-8f43-b423714861ec",
		Name:                  "Containment Membrane",
		Completeness:          effects.CompletenessFull,
		AlternativeCosts:      []game.AlternativeCost{effects.Surge("{U}")},
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
	})
}
