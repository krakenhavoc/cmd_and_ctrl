package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Praetor's Counsel — Sorcery {5}{G}{G}{G}:
//
//	"Return all cards from your graveyard to your hand. Exile Praetor's
//	 Counsel. You have no maximum hand size for the rest of the game."
//
// The graveyard is read once before anything moves, and the spell is not
// in it: it is still on the stack while it resolves, so it neither returns
// itself nor needs a filter. Praetor's Counsel then exiles itself (the
// resolving card moving itself, as Finale of Revelation does, CR 608.2n),
// and the engine skips the ordinary graveyard routing.
//
// "No maximum hand size for the rest of the game" is the player-level
// grant (Game.SetMaxHandSizeForEffect), not a static of a permanent: the
// spell is in exile a moment after it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83d8eda7-b8d2-479a-b1ec-de56833a3182",
		Name:         "Praetor's Counsel",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := returnWholeGraveyardToHand(ctx, func(game.Card) bool { return true }); err != nil {
				return err
			}
			if err := (ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.SetMaxHandSizeForEffect(ctx.Controller(), game.NoMaxHandSize)
		},
	})
}

// returnWholeGraveyardToHand is "return all <matching> cards from your
// graveyard to your hand": the cards are read once, then each moves.
func returnWholeGraveyardToHand(ctx *Context, match func(game.Card) bool) error {
	for _, id := range graveyardCardIDs(ctx, ctx.Controller(), match) {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
