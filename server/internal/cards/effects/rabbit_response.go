package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabbit Response — Instant {2}{W}{W} (Bloomburrow, #4467):
//
//	"Creatures you control get +2/+1 until end of turn. If you control a
//	 Rabbit, scry 2."
//
// The pump is Makindi Stampede's: the creatures you control as it
// resolves, locked in for the turn (CR 611.2c), so one that enters later
// gets nothing. "If you control a Rabbit" is read after the pump, as the
// spell resolves, over any permanent you control with the subtype (read
// through this card's text).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f7abd7d-1087-495c-8a52-2a2368cc1032",
		Name:         "Rabbit Response",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump}, // the scry rides on a Rabbit, so it is not declared
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 1, Label: "Rabbit Response — +2/+1"},
			Instruction(func(ctx *Context) error {
				if len(MatchingBattlefield(ctx, And(Subtype("Rabbit"), YouControl()))) == 0 {
					return nil
				}
				return Scry{N: 2}.Apply(ctx)
			}),
		),
	})
}
