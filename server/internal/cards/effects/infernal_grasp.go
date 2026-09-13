package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infernal Grasp — Instant {1}{B} (EDHREC rank 261):
//
//	"Destroy target creature. You lose 2 life."
//
// Black's two-mana unconditional answer; the life is the price. The
// life loss is not damage — no prevention, no doubling. A target
// that left in response fizzles the whole spell (CR 608.2b: one
// target, all illegal) and nobody loses anything, which is the
// printed behaviour.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94f0a572-e91c-4b56-a5d1-6cbbeabd210d",
		Name:         "Infernal Grasp",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
