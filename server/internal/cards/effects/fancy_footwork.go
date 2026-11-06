package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fancy Footwork — Instant — Lesson {2}{W}:
//
//	"Untap one or two target creatures. They each get +2/+2 until end
//	 of turn."
//
// Each target still legal on resolution is untapped and pumped; one
// that became illegal is neither (CR 608.2b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5648a32-98b0-4625-aae1-ae56e0ef0197",
		Name:         "Fancy Footwork",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := untilEndOfTurn(ctx, t.ID, nil, "Fancy Footwork — +2/+2", game.ModifyPTMod(2, 2)); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
