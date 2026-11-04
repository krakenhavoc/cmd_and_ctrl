package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerial Assault — Sorcery {2}{W} (Jumpstart):
//
//	"Destroy target tapped creature. You gain 1 life for each creature you
//	 control with flying."
//
// Royal Assassin's "target tapped creature", re-checked on resolution
// (CR 608.2b): a creature untapped in response makes the spell do nothing,
// the life included. The fliers are counted after the destruction (CR
// 608.2c, in printed order), so a flier of yours that was the target no
// longer counts; one that survived the destruction does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4eb06ca-16d8-41c2-8fc6-22e73872b5b0",
		Name:         "Aerial Assault",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target tapped creature", tappedPermanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Amount: jmpB7CreaturesYouControlWithFlying(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
