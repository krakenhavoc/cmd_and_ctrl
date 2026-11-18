package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunset Revelry — Sorcery {1}{W} (Midnight Hunt, #5497):
//
//	"If an opponent has more life than you, you gain 4 life.
//	 If an opponent controls more creatures than you, create two 1/1 white
//	 Human creature tokens.
//	 If an opponent has more cards in hand than you, draw a card."
//
// Beza's catch-up clauses (b39AnOpponentHasMoreLife and its siblings), each an
// independent "some single opponent beats you" test that is strictly more,
// never tied, and skips an eliminated seat (CR 800.4a). They are carried out
// in printed order as each is reached (CR 608.2): the creature count is made
// before the Humans exist and the hand count before the draw. The spell has
// already left the hand when it resolves, so it is not counted in the hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65779f3b-7fc5-4847-99d7-78f697d34c3d",
		Name:         "Sunset Revelry",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1}, // the third clause
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g, you := ctx.Game, item.Controller
			if b39AnOpponentHasMoreLife(g, you) {
				if err := (GainLife{Player: you, Amount: 4}).Apply(ctx); err != nil {
					return err
				}
			}
			if b39AnOpponentControlsMoreCreatures(g, you) {
				if err := (CreateToken{Controller: you, Template: TokenCard("1/1 white Human"), N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			if b39AnOpponentHasMoreCardsInHand(g, you) {
				return DrawCards{Player: you, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
