package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hold for Questioning — Enchantment — Aura {3}{U} (Jumpstart 2022, #5948):
//
//	"Enchant creature or planeswalker
//	 When this Aura enters, tap enchanted permanent and investigate.
//	 Enchanted permanent doesn't untap during its controller's untap step and
//	 its activated abilities can't be activated."
//
// Arrest's activation bits on a creature or planeswalker (CR 602.5, loyalty
// abilities included, CR 606.1), Castaway's Despair's untap hold, and an
// entry trigger that taps what the Aura enchants and then investigates for
// the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "1095fcd9-8127-4345-9052-fccb350296d3",
		Name:                  "Hold for Questioning",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.TargetPermanent("enchant creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantActivate | game.CantActivateMana),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hold for Questioning — tap enchanted permanent and investigate", effects.GapB169TapEnchantedAndInvestigate),
		},
	})
}
