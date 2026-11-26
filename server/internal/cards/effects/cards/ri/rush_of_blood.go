package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rush of Blood — {2}{R} Instant: "Target creature gets +X/+0 until end of turn,
// where X is its power."
//
// X is the target's power as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0bacbf1-3c35-436e-b0bf-10613b60b726",
		Name:         "Rush of Blood",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found || c.CurrentPower() == 0 {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: c.CurrentPower(), Label: "Rush of Blood — +X/+0"}.Apply(ctx)
		},
	})
}
