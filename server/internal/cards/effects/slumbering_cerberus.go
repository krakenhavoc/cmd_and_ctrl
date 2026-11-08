package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slumbering Cerberus — Creature — Dog {1}{R}, 4/2 (Foundations, #4465):
//
//	"This creature doesn't untap during your untap step.
//	 Morbid — At the beginning of each end step, if a creature died this
//	 turn, untap this creature."
//
// Marjhan's untap-step restriction (CR 502.3) and Zombie Ogre's morbid
// intervening "if" (CR 603.4, condACreatureDiedThisTurn): any creature,
// anyone's, tokens included, checked as the end step begins and again as
// the ability resolves. It triggers in every player's end step, not only
// its controller's. The untap is of this object: a Cerberus that left and
// came back is a new object and is not untapped (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "39cae244-96e1-4ea3-899a-3aa871cca31c",
		Name:                  "Slumbering Cerberus",
		Completeness:          CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				On(game.EventBeginEndStep, AnyPlayer, "Slumbering Cerberus — morbid: untap this creature",
					thisStillHere(b31UntapSelf))),
		},
	})
}
