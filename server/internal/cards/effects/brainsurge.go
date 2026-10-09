package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brainsurge — Instant {2}{U}:
//
//	"Draw four cards, then put two cards from your hand on top of your
//	 library in any order."
//
// Brainstorm with a bigger draw. See brainstorm.go for why the put-back
// is raised after the draw (CR 608.2c) and how the order is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46c727cb-1f47-4775-8af4-0230ef53966b",
		Name:         "Brainsurge",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 4, game.DrawThen{
				Ref: rfReprintABrainsurgePutBack, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}
