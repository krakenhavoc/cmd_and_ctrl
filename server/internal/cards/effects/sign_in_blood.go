package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sign in Blood — "Target player draws two cards and loses 2 life."
// Composition of DrawCards + ChangePlayerLife applied to the chosen
// player. Canonical multi-player effect: the caster can draw off
// themselves or punish an opponent's hand.
func init() {
	Register(Spec{
		OracleID: "c6207f6a-a624-4754-88f5-dbe700c841ff",
		Name:     "Sign in Blood",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			target := item.Targets[0].ID
			if err := (DrawCards{Player: target, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), target, -2)
		},
	})
}
