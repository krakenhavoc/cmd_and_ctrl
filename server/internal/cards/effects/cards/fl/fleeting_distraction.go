package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleeting Distraction — Instant {U} (Foundations, #4465):
//
//	"Target creature gets -1/-0 until end of turn.
//	 Draw a card."
//
// Rebellious Strike's shape with a shrink: the -1/-0, then the draw in
// printed order. A spell whose only target is illegal as it resolves does
// nothing at all (CR 608.2b), so the card is drawn only when the shrink
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b36271cb-cda2-434a-a580-6050dc460409",
		Name:         "Fleeting Distraction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.HarmUntilEOT(0, -1, 0)), Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.BoostTheTargetUntilEOT(-1, 0, "Fleeting Distraction — -1/-0")(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
