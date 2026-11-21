package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildsize — Instant {2}{G} (Ravnica: Clue Edition):
//
//	"Target creature gets +2/+2 and gains trample until end of turn.
//	 Draw a card."
//
// Massive Might's body, then the controller draws. A target that has become
// illegal counters the whole spell, draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1f439c7-217f-4b20-a77d-b9014c5de8c0",
		Name:         "Wildsize",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PumpAndGrantFirstTarget(2, 2, "Wildsize — +2/+2 and trample", "trample")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
