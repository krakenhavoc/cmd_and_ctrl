package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9f7abd7d-1087-495c-8a52-2a2368cc1032",
		Name:         "Rabbit Response",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump}, // the scry rides on a Rabbit, so it is not declared
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Toughness: 1, Label: "Rabbit Response — +2/+1"},
			effects.Instruction(func(ctx *effects.Context) error {
				if len(effects.MatchingBattlefield(ctx, effects.And(effects.Subtype("Rabbit"), effects.YouControl()))) == 0 {
					return nil
				}
				return effects.Scry{N: 2}.Apply(ctx)
			}),
		),
	})
}
