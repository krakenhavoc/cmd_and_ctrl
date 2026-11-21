package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectacular Tactics — Instant {1}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"Choose one —
//	 • Put a +1/+1 counter on target creature you control. It gains
//	   hexproof until end of turn.
//	 • Destroy target creature with power 4 or greater."
//
// Each bullet carries its own target clause. The first is two steps on
// its one target (the counter, then the keyword); the second is the
// shared DestroyTheModesTarget.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15948fdf-3f9e-4eed-b85b-c5a554ff0571",
		Name:         "Spectacular Tactics",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Put a +1/+1 counter on target creature you control. It gains hexproof until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := effects.PlusOneCounterOnTheModesTarget(nil, ctx, occ); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Spectacular Tactics — hexproof"}.Apply(ctx)
				}),
			effects.ModeDoing("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
				effects.DestroyTheModesTarget),
		),
	})
}
