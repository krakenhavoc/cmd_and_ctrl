package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hamato Guardian Stance — {W} Instant (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Target creature gets +1/+3 and gains flying until end of turn.
//	 Scry 1."
//
// Titan's Strength's shape: the pump and the flying on the still-legal
// target (pumpAndGrantFirstTarget), then the scry, which is last so
// nothing waits on its prompt (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e08b19ca-f9d7-4c33-b37a-a77b37aabf1b",
		Name:         "Hamato Guardian Stance",
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "flying")), game.Purpose{Scry: 1}),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PumpAndGrantFirstTarget(1, 3, "Hamato Guardian Stance — +1/+3 and flying", "flying")(ctx.Game, item); err != nil {
				return err
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
