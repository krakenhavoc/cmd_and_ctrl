package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rebellious Strike — Instant {1}{W} (Tarkir: Dragonstorm, #4463):
//
//	"Target creature gets +3/+0 until end of turn.
//	 Draw a card."
//
// Sugar Rush's shape in white: the pump, then the draw in printed order. A
// spell whose only target is illegal as it resolves does nothing at all
// (CR 608.2b), so the card is drawn only when the pump happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f3d7102-f8ce-42d4-94eb-f69b695773d1",
		Name:         "Rebellious Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 3, 0)), Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(3, 0, "Rebellious Strike — +3/+0")(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
