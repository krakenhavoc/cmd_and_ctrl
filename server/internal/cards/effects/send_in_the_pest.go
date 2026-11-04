package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Send in the Pest — Sorcery {1}{B} (Secrets of Strixhaven, #4456):
//
//	"Each opponent discards a card. You create a 1/1 black and green Pest
//	 creature token with "Whenever this token attacks, you gain 1 life.""
//
// The discard is every opponent choosing their own card at once
// (eachOpponentDiscardsOne); the token is the Secrets of Strixhaven Pest
// (PestAttackToken) and does not wait on the discards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "955bfcb8-6477-4ed0-b6e7-8940874ae5e4",
		Name:         "Send in the Pest",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := eachOpponentDiscardsOne(ctx.Game, item); err != nil {
				return err
			}
			return CreateToken{Template: PestAttackToken(), N: 1}.Apply(ctx)
		},
	})
}
