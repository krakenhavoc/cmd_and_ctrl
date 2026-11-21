package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amazing Acrobatics — Instant {1}{U}{U} (Marvel's Spider-Man):
//
//	"Choose one or both —
//	 • Counter target spell.
//	 • Tap one or two target creatures."
//
// Brutal Expulsion's "choose one or both" (CR 700.2): each bullet reads
// its own targets, in printed order (CR 608.2c). A spell that has left
// the stack, or a creature that is no longer a legal target, is skipped
// (CR 608.2b), so the other bullet still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0466f23c-d07c-4015-82d5-eab4b226f1ff",
		Name:         "Amazing Acrobatics",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Counter target spell.", effects.TargetSpell("target spell"), effects.CounterTheModesTarget),
			effects.ModeDoing("Tap one or two target creatures.", effects.TargetCreature("one or two target creatures").WithCount(1, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		),
	})
}
