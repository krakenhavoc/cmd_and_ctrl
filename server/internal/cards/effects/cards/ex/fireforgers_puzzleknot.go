package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fireforger's Puzzleknot — Artifact {2} (Kaladesh, #5539):
//
//	"When this artifact enters, it deals 1 damage to any target.
//	 {2}{R}, Sacrifice this artifact: It deals 1 damage to any target."
//
// The entry trigger is Karplusan Hound's targeted damage from the source
// (target chosen as the trigger goes on the stack, checked again as it
// resolves). The activated ability is Springjaw Trap's sacrifice-as-a-cost
// damage: the damage still comes from the Puzzleknot, read from last-known
// information after it has been sacrificed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0de18ba2-84a7-4e07-84b2-771761d77aaf",
		Name:         "Fireforger's Puzzleknot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WhenThisEnters("Fireforger's Puzzleknot — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
					effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice this artifact: It deals 1 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
	})
}
