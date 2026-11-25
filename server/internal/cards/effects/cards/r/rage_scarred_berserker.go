package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rage-Scarred Berserker — Creature — Minotaur Berserker {4}{B}, 5/4 (Theros Beyond Death, #5514):
//
//	"When this creature enters, target creature you control gets +1/+0 and gains
//	 indestructible until end of turn."
//
// Devotion is counted as the trigger resolves (CR 700.5, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "658d54d4-e967-490c-b958-3743b3b1ef45",
		Name:         "Rage-Scarred Berserker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Rage-Scarred Berserker — target creature you control gets +1/+0 and gains indestructible until end of turn",
				effects.PumpAndGrantFirstTarget(1, 0, "Rage-Scarred Berserker — +1/+0 and indestructible until end of turn", "indestructible")),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible"))),
		},
	})
}
