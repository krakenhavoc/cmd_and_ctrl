package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chilling Trap — {U} Instant:
//
//	"Target creature gets -4/-0 until end of turn. If you control a
//	Wizard, draw a card."
//
// The Wizard check is made as the spell resolves (CR 608.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ac49129-7da2-4729-b48c-b6da279c212e",
		Name:         "Chilling Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -4, Label: "Chilling Trap — -4/-0"}).Apply(ctx); err != nil {
					return err
				}
			}
			if effects.ControlsA(ctx.TextSubtype("Wizard"))(ctx.Game, ctx.Controller()) {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
