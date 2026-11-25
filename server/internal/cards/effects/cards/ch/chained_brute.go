package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chained Brute — Creature — Devil {1}{R}, 4/3 (Jumpstart, #5510):
//
//	"This creature doesn't untap during your untap step.
//	 {1}, Sacrifice another creature: Untap this creature. Activate only
//	 during your turn."
//
// Battered Golem's untap restriction and Fain's untap-this body; the
// sacrifice is of another creature, so the Brute itself is never a payment.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "3835cb81-b6e4-4be8-9d49-8307ddc2e1a2",
		Name:                  "Chained Brute",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}, Sacrifice another creature: Untap this creature. Activate only during your turn.",
			Cost:      effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Condition: effects.DuringYourTurn(),
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet},
			Effect:    effects.B31UntapSelf,
		}},
	})
}
