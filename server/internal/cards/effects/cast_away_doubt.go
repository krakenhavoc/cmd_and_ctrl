package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cast Away Doubt — Sorcery {2}{B}:
//
//	"Draw two cards. Cast Away Doubt deals 2 damage to each player."
//
// The draw comes first, as printed. The damage is one instance dealt to
// every seat still in the game, the caster included, by the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bcf29dd7-6d5a-4139-958e-7c59434c0770",
		Name:         "Cast Away Doubt",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return b38DamageEachPlayer(ctx, 2)
		},
	})
}
