package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invoke the Divine — Instant {2}{W}:
//
//	"Destroy target artifact or enchantment. You gain 4 life."
//
// A target that is illegal on resolution fizzles the whole spell, so no
// life is gained (the ruling, CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "87bd0fdb-c5b9-46ea-9858-1a870960351f",
		Name:         "Invoke the Divine",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
		},
	})
}
