package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Withering Torment — Instant {2}{B} (EDHREC rank 264):
//
//	"Destroy target creature or enchantment. You lose 2 life."
//
// Infernal Grasp that also answers an enchantment, for one more
// mana. Same shape: destroy, then the controller loses 2 life —
// loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "ffce81c5-1b58-4882-a4e7-6f8d7cb170de",
		Name:     "Withering Torment",
		Targets:  TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
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
