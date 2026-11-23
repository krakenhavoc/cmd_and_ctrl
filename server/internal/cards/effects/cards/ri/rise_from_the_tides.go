package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise from the Tides — Sorcery {5}{U} (Innistrad Remastered):
//
//	"Create a tapped 2/2 black Zombie creature token for each instant and
//	 sorcery card in your graveyard."
//
// Counted as the spell resolves; the spell is on the stack, not in the
// graveyard, so it is not counted. With none, no token is made.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "390b862a-2a85-44bd-832c-f5bbd3eb4ea0",
		Name:         "Rise from the Tides",
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.GapGraveyardCardsMatching(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
			if n == 0 {
				return nil
			}
			return effects.CreateTokenAdvanced{
				Controller: ctx.Controller(),
				Spec:       effects.Token(effects.BlackZombieToken()).EntersTapped(),
				N:          n,
			}.Apply(ctx)
		},
	})
}
