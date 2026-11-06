package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eriette's Lullaby — Sorcery {1}{W} (Outlaws of Thunder Junction, #4468):
//
//	"Destroy target tapped creature. You gain 2 life."
//
// Sheer Drop's "target tapped creature", then Venom's Hunger's life. The
// target is checked again on resolution (CR 608.2b): a creature that untapped
// in response is no longer legal, and with no legal target the whole spell
// is countered, so no life is gained. A legal target that survives the
// destruction (indestructible) still gives you the 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b134a9e6-b004-4833-b298-10529f5a85b8",
		Name:         "Eriette's Lullaby",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
