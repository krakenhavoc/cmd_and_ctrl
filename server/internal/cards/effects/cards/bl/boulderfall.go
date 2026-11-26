package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boulderfall — Instant {6}{R}{R} (Theros):
//
//	"Boulderfall deals 5 damage divided as you choose among any number of
//	 targets."
//
// Bogardan Hellkite's division: announced with the targets (CR 601.2d), each
// target getting at least 1; a target that is illegal on resolution takes
// nothing and its share is lost (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2160aed4-5a7a-4eca-96b5-707204fa73bb",
		Name:         "Boulderfall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(5)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
