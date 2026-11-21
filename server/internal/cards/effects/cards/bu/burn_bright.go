package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burn Bright — Instant {2}{R} (Ravnica Allegiance, #5521):
//
//	"Creatures you control get +2/+0 until end of turn."
//
// Path of Anger's Flame's team pump, at instant speed: the set is the
// creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed14c7fa-632c-47ee-9db0-8c492771121a",
		Name:         "Burn Bright",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: effects.ResolveSteps(effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2,
			Label: "Burn Bright — +2/+0",
		}),
	})
}
