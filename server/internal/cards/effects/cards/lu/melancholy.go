package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Melancholy — Enchantment — Aura {2}{B} (Planar Chaos, #5595):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 At the beginning of your upkeep, sacrifice this Aura unless you pay {B}."
//
// Bitter Chill's tap-and-hold (tapEnchantedCreatureOnEntry and
// enchantedDoesntUntap) with the Elder Dragons' upkeep tax
// (UpkeepSacrificeUnlessYouPay): declining, or a payment the controller
// can't fund, sacrifices the Aura and frees the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "58cc7682-f8ce-47cf-a291-0301fea9ddfd",
		Name:                  "Melancholy",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Melancholy — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
			effects.UpkeepSacrificeUnlessYouPay("Melancholy", "{B}"),
		},
	})
}
