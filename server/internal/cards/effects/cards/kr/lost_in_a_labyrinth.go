package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost in a Labyrinth — Instant {U} (Theros, S82 OLD b389, #5561):
//
//	"Target creature gets -3/-0 until end of turn. Scry 1."
//
// Befuddle's shrink with Furious Bellow's scry (CR 701.22a).
//
// No simplification.
func init() {
	purpose := effects.ForTargets(effects.HarmUntilEOT(0, -3, 0))
	purpose.Scry = 1
	effects.Register(effects.Spec{
		OracleID:     "e6813885-b629-4bd4-b445-da8ab4363080",
		Name:         "Lost in a Labyrinth",
		Completeness: effects.CompletenessFull,
		Purpose:      purpose,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -3, Label: "Lost in a Labyrinth — -3/-0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
