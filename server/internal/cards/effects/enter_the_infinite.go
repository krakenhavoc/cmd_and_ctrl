package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enter the Infinite — Sorcery {8}{U}{U}{U}{U}:
//
//	"Draw cards equal to the number of cards in your library, then put
//	 a card from your hand on top of your library. You have no maximum
//	 hand size until your next turn."
//
// The count is read as the spell resolves (CR 608.2h), so the draw
// takes exactly the library and nobody loses to an empty draw. The
// put-back is Brainstorm's, for one card. "Until your next turn" is a
// granted no-maximum entry with ADR 0063's until-your-next-turn
// duration (game.GrantHandSizeForEffect, #2108), written before the
// draw: nothing in the draw reads the maximum, and it keeps the grant
// from waiting on a prompt the draw pauses on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2adbb56a-45e9-4fbe-b586-3488ef8014a3",
		Name:         "Enter the Infinite",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := ctx.Game.GrantHandSizeForEffect(item.Controller, game.HandSizeNoMaximum, 0,
				"Enter the Infinite — no maximum hand size", item.SourceCardID,
				DurationUntilYourNextTurn(ctx, item.Controller)); err != nil {
				return err
			}
			n := 0
			if p := ctx.PlayerByID(item.Controller); p != nil && p.Library != nil {
				n = p.Library.Size()
			}
			return ctx.Game.DrawNThenForEffect(item.Controller, n, game.DrawThen{
				Ref: drawThenEnterTheInfinitePutBack, Player: item.Controller, Source: item.SourceCardID, N: 1,
			})
		},
	})
}
