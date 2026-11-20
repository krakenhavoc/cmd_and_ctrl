package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fated Intervention — Instant {2}{G}{G}{G} (Born of the Gods,
// #5559):
//
//	"Create two 3/3 green Centaur enchantment creature tokens. If it's
//	 your turn, scry 2."
//
// Two tokens from the table, then the scry behind "if it's your turn"
// (IsYourTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f85342b-b55b-461b-a3dc-17815e02d04d",
		Name:         "Fated Intervention",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2, Scry: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: TokenCard("3/3 green Centaur enchantment"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			if !IsYourTurn(ctx.Game, item.Controller) {
				return nil
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
