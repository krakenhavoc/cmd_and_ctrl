package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minions' Murmurs — Sorcery {2}{B}{B} (Future Sight, #5594):
//
//	"You draw X cards and you lose X life, where X is the number of
//	 creatures you control."
//
// X is counted once, as the spell resolves, and both halves use it
// (Champion of Dusk's shape over every creature, not one type). With no
// creature, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd755ec1-6727-4165-a7fd-5d9b59754ada",
		Name:         "Minions' Murmurs",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsCreature() })
			if x <= 0 {
				return nil
			}
			if err := (DrawCards{Player: item.Controller, N: x}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
		},
	})
}
