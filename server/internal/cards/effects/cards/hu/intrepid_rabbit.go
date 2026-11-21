package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intrepid Rabbit — Creature — Rabbit Soldier {2}{W}, 3/2 (Bloomburrow,
// #4467):
//
//	"Offspring {1} (You may pay an additional {1} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 When this creature enters, target creature you control gets +1/+1
//	 until end of turn."
//
// Offspring is the shared keyword (offspring.go). The entry pump is
// Shrouded Shepherd's targeted pump at +1/+1, chosen as the trigger goes
// on the stack; the Rabbit may target itself. The offspring token is a
// copy, so it has the entry pump too and pumps a creature as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7ed4afc9-959a-4868-8f4a-2221bea1ba32",
		Name:          "Intrepid Rabbit",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Offspring("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.OffspringToken("Intrepid Rabbit"),
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Intrepid Rabbit — target creature you control gets +1/+1 until end of turn",
				effects.BoostTargetUntilEOT("Intrepid Rabbit — +1/+1", 1, 1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
		},
	})
}
