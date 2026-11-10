package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gnawing Crescendo — Instant {2}{R} (Wilds of Eldraine, #4471):
//
//	"Creatures you control get +2/+0 until end of turn. Whenever a
//	 nontoken creature you control dies this turn, create a 1/1 black Rat
//	 creature token with "This token can't block.""
//
// The pump is Rabbit Response's: the creatures you control as the spell
// resolves, locked in for the turn (CR 611.2c). The second sentence is a
// repeating delayed triggered ability (CR 603.7b, Benefactor's Draught's
// shape) that lasts until cleanup: it fires for each nontoken creature
// that dies under the caster's control, judged on the creature as it last
// existed (CR 603.10a, Blight Mound's reading), so a creature that stopped
// being a creature or changed hands before it died is read as it was. The
// Rat is Twisted Sewer-Witch's printed token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0332f7b5-dad3-4fd0-b86c-78bd8301f59d",
		Name:         "Gnawing Crescendo",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Gnawing Crescendo — +2/+0"},
			Instruction(func(ctx *Context) error {
				return DelayedOnEvent{
					Label:     "Gnawing Crescendo — create a 1/1 black Rat that can't block",
					On:        []game.EventKind{game.EventLTB},
					Condition: std54NontokenCreatureYouControlDiedCondition,
					Body:      std54CantBlockRatBody,
					Repeats:   true,
				}.Apply(ctx)
			}),
		),
	})
}
