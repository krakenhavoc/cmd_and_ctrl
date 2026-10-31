package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "15948fdf-3f9e-4eed-b85b-c5a554ff0571",
		Name:         "Spectacular Tactics",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Put a +1/+1 counter on target creature you control. It gains hexproof until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := PlusOneCounterOnTheModesTarget(nil, ctx, occ); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Spectacular Tactics — hexproof"}.Apply(ctx)
				}),
			ModeDoing("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4)),
				DestroyTheModesTarget),
		),
	})
}
