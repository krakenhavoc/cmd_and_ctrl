package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plague of Vermin — Sorcery {6}{B} (Mirage, #5908):
//
//	"Starting with you, each player may pay any amount of life. Repeat
//	 this process until no one pays life. Each player creates a 1/1
//	 black Rat creature token for each 1 life they paid this way."
//
// Each player is asked in turn order from the controller (the pay-any-
// amount prompt, ADR 0129), capped at their life total (CR 119.4). A pass
// in which anyone paid is followed by another; a pass in which nobody paid
// ends it, and then every player makes their own Rats for the total they
// paid across all passes. The running totals ride the prompts as data.
func init() {
	Register(Spec{
		OracleID:     "efc12acf-ffef-40a6-b2ce-fe93e89ad840",
		Name:         "Plague of Vermin",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			order := plagueOfVerminOrder(ctx.Game, ctx.Controller())
			return plagueOfVerminAsk(ctx, order, make([]int, 2+len(order)))
		},
	})
}
