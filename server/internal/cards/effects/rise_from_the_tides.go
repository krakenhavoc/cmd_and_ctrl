package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise from the Tides — Sorcery {5}{U} (Innistrad Remastered):
//
//	"Create a tapped 2/2 black Zombie creature token for each instant and
//	 sorcery card in your graveyard."
//
// Counted as the spell resolves; the spell is on the stack, not in the
// graveyard, so it is not counted. With none, no token is made.
func init() {
	Register(Spec{
		OracleID:     "390b862a-2a85-44bd-832c-f5bbd3eb4ea0",
		Name:         "Rise from the Tides",
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := gapGraveyardCardsMatching(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
			if n == 0 {
				return nil
			}
			return CreateTokenAdvanced{
				Controller: ctx.Controller(),
				Spec:       Token(BlackZombieToken()).EntersTapped(),
				N:          n,
			}.Apply(ctx)
		},
	})
}
