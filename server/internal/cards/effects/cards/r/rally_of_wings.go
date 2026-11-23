package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally of Wings — Instant {1}{W} (Foundations Jumpstart):
//
//	"Untap all creatures you control. Creatures you control with flying
//	 get +2/+2 until end of turn."
//
// The untap first, then the pump (CR 608.2c). The pumped set is the
// creatures you control with flying as the spell resolves, read through
// the layers and locked then (CR 611.2c), so a creature that gains flying
// later this turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "984336a7-85bd-4549-9771-03a0c2367db0",
		Name:         "Rally of Wings",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: effects.ResolveSteps(
			effects.UntapAllCreaturesYouControl{},
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("flying")),
					Power:     2,
					Toughness: 2,
					Label:     "Rally of Wings — +2/+2",
				}.Apply(ctx)
			}),
		),
	})
}
