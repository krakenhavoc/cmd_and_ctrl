package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Return to the Light Realms — Sorcery {7}{W}{W} (Reality Fracture, tracker #2795):
//
//	"Return all nonland permanent cards from your graveyard to the
//	 battlefield."
//
// The set is read once before anything moves, and the cards enter
// together (CR 603.6a) under your control. No target, so nothing can be
// answered by removing a target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "217e76ca-ea94-417c-b5b0-82e2302eeee8",
		Name:         "Return to the Light Realms",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			match := NonlandPermanentCard()
			return ReturnFromGraveyardTogether{
				Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return match(ctx.Game, ctx.Controller(), c)
				}),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
