package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ballista Charger — Artifact — Vehicle {5}, 6/6 (Kaladesh, #5539):
//
//	"Whenever this Vehicle attacks, it deals 1 damage to any target.
//	 Crew 3"
//
// Karplusan Hound's attack trigger without the intervening "if": the target is
// chosen as the trigger goes on the stack and checked again as it resolves
// (CR 608.2b), the damage coming from the Vehicle. Crew 3 is the ordinary crew
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4862430-e3d3-4f29-90ca-50141e36e2f1",
		Name:         "Ballista Charger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WheneverThisAttacks("Ballista Charger — 1 damage to any target", effects.SourceDealsDamageToEachLegalTarget(1)),
					effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Ballista Charger"),
		}},
	})
}
