package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "34237f39-1db3-4b3d-b87b-5fd1fbd5462a",
		Name:         "Wolf Strike",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ItsNight(ctx.Game) {
				if biter, ok := ctx.ClauseTarget(0); ok && biter.Kind == game.TargetCard {
					if err := (BoostUntilEOT{Target: biter.ID, Power: 2, Label: "Wolf Strike — +2/+0 at night"}).Apply(ctx); err != nil {
						return err
					}
				}
				// The bite reads the biter's power as it now stands.
				ctx.Game.RecomputeLayersIfStaleLocked()
			}
			return oneSidedBite(item, ctx)
		},
	})
}
