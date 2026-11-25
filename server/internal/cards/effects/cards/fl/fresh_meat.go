package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fresh Meat — Instant {3}{G} (Commander 2019):
//
//	"Create a 3/3 green Beast creature token for each creature put into your
//	 graveyard from the battlefield this turn."
//
// Caller of the Claw's count with the tokens kept in: every creature, token or
// not, that went from the battlefield to your graveyard this turn, read from
// the turn's event log as the spell resolves and counted once per trip, a
// creature that has since left the graveyard included (Fresh Meat's own count
// is what the creatures "were put", not what they are).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3be6e78f-f9ec-4bf4-b70f-40ba677aaa48",
		Name:         "Fresh Meat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := effects.OldB152CreaturesPutIntoYourGraveyardFromTheBattlefieldThisTurn(ctx.Game, ctx.Controller())
			if n <= 0 {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: n}.Apply(ctx)
		},
	})
}
