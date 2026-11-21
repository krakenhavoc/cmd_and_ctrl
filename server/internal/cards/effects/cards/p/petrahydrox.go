package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "545c516f-85a4-4a59-8cfa-03da27121a57",
		Name:         "Petrahydrox",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self,
				"Petrahydrox — return this creature to its owner's hand",
				effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
		},
	})
}
