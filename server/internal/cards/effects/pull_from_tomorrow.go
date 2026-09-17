package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pull from Tomorrow — Instant for {X}{U}{U}:
//
//	"Draw X cards, then discard a card."
//
// The discard is a cost of doing business on every other deck and a
// bonus on this one. Printed order matters and the engine keeps it:
// the discard choice is queued against the post-draw hand, so a card
// just drawn is a legal pitch.
func init() {
	Register(Spec{
		OracleID: "b1a23235-3076-475c-a68a-db29cf2a9dba",
		Name:     "Pull from Tomorrow",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: ctx.X()}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: item.Controller,
				Source: item.SourceCardID,
				N:      1,
			})
			return nil
		},
	})
}
