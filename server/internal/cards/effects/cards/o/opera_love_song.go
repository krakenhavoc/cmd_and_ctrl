package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "22f259b3-9de1-44a1-a89d-f2f1b1249d4d",
		Name:         "Opera Love Song",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Exile the top two cards of your library. You may play those cards until your next end step."),
			effects.Mode("One or two target creatures each get +2/+0 until end of turn.",
				effects.TargetCreature("one or two target creatures").WithCount(1, 2)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.ExileTopNUntilYourNextEndStep(ctx, 2)
			}
			for _, t := range ctx.ModeTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Opera Love Song — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
