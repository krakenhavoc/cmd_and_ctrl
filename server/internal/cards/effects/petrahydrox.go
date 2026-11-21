package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Petrahydrox — Creature — Weird {3}{U/R}, 3/3 (Ravnica Remastered, #5948):
//
//	"When this creature becomes the target of a spell or ability, return this
//	 creature to its owner's hand."
//
// Illusionary Servant's trigger with Fleeting Effigy's return: it fires when
// the Petrahydrox becomes the target of any spell or ability, whoever controls
// it, goes on the stack above it and resolves first, so the spell or ability
// is left without a target (CR 608.2b). It returns only the permanent the
// ability came from (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "545c516f-85a4-4a59-8cfa-03da27121a57",
		Name:         "Petrahydrox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self,
				"Petrahydrox — return this creature to its owner's hand",
				Do(Instruction(stdB9ReturnThisToHand))),
		},
	})
}
