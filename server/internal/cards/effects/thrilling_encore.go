package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrilling Encore — Instant {4}{B} (Wilds of Eldraine Commander):
//
//	"Put onto the battlefield under your control all creature cards in all
//	 graveyards that were put there from the battlefield this turn."
//
// Every graveyard is read as the spell resolves. A card counts when its
// newest arrival in its graveyard this turn came from the battlefield
// (putIntoGraveyardFromBattlefieldThisTurn, Twilight Shepherd's read; a card
// that left and came back is a new object, CR 400.7, and a token that died
// ceased to exist). Tokens' cards, creatures milled or discarded, and
// creatures from earlier turns stay put. They all enter under your control in
// one simultaneous entry (CR 614.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2a114ff8-fa7d-41b8-b412-6232fbb68efc",
		Name:         "Thrilling Encore",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets: allGraveyardsCardIDs(ctx, func(c game.Card) bool {
					return c.IsCreature() && putIntoGraveyardFromBattlefieldThisTurn(ctx.Game, c.InstanceID)
				}),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
