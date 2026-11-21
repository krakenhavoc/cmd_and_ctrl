package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasp of Darkness — Instant {B}{B} (Scars of Mirrodin, #5577):
//
//	"Target creature gets -4/-4 until end of turn."
//
// Disfigure with a bigger number.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "494c819d-7d73-432b-8c96-4cb9dcd31094",
		Name:         "Grasp of Darkness",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Grasp of Darkness — -4/-4"}.Apply(ctx)
		},
	})
}
