package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commune with Lava — Instant {X}{R}{R}:
//
//	"Exile the top X cards of your library. Until the end of your next
//	 turn, you may play those cards."
//
// Alania's Pathmaker's window (the end of the controller's next turn) over
// X cards. The permission is "play", so a land among them may be played
// with a land drop (CastOnly is left off). An X of 0 exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4aa0f64-d202-42a1-af62-0e53bbe87782",
		Name:         "Commune with Lava",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			window := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
			return effects.ExileTopWithPermission{
				From:    ctx.Controller(),
				GrantTo: ctx.Controller(),
				N:       x,
				Window:  &window,
			}.Apply(ctx)
		},
	})
}
