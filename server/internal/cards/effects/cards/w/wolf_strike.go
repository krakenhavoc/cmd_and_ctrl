package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolf Strike — Instant {2}{G} (#2586, ADR 0132):
//
//	"Target creature you control gets +2/+0 until end of turn if it's
//	 night. Then it deals damage equal to its power to target creature
//	 you don't control."
//
// Bite Down's two-clause one-sided fight with a night rider in front.
// "If it's night" is read as the spell resolves (ItsNight), and the
// biter's power is read AFTER the boost, which is the printed order.
// Either creature having left leaves the other half undone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34237f39-1db3-4b3d-b87b-5fd1fbd5462a",
		Name:         "Wolf Strike",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.ItsNight(ctx.Game) {
				if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
					if err := (effects.BoostUntilEOT{Target: biter.ID, Power: 2, Label: "Wolf Strike — +2/+0 at night"}).Apply(ctx); err != nil {
						return err
					}
				}
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
			}
			return effects.OneSidedBite(item, ctx)
		},
	})
}
