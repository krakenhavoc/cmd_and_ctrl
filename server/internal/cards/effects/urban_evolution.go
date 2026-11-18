package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urban Evolution — Sorcery {3}{G}{U} (S82 reprint pass):
//
//	"Draw three cards. You may play an additional land this turn."
//
// The extra land drop is a one-shot grant for this turn (Explore's seam).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c715e525-fc3a-4cab-94a0-04e0a22a5013",
		Name:         "Urban Evolution",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}
