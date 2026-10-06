package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Opera Love Song — Instant {1}{R}:
//
//	"Choose one —
//	 • Exile the top two cards of your library. You may play those
//	   cards until your next end step.
//	 • One or two target creatures each get +2/+0 until end of turn."
//
// Mode 0 is game.UntilYourNextEndStep (#2373). Mode 1 locks its
// affected set when the spell resolves (CR 611.2c) and skips a target
// that left in response.
func init() {
	Register(Spec{
		OracleID:     "22f259b3-9de1-44a1-a89d-f2f1b1249d4d",
		Name:         "Opera Love Song",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Exile the top two cards of your library. You may play those cards until your next end step."),
			Mode("One or two target creatures each get +2/+0 until end of turn.",
				TargetCreature("one or two target creatures").WithCount(1, 2)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return ExileTopNUntilYourNextEndStep(ctx, 2)
			}
			for _, t := range ctx.ModeTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Label: "Opera Love Song — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
