package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Survey the Wreckage — Sorcery {4}{R} (Return to Ravnica, #5566):
//
//	"Destroy target land. Create a 1/1 red Goblin creature token."
//
// One target clause. When the land is an illegal target on resolution the whole
// spell does nothing (CR 608.2b), the token included; when it is legal but
// survives (indestructible) the token is still made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35df2d01-2f9f-49a7-8fa1-cfaa52e5c7b6",
		Name:         "Survey the Wreckage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.Do(effects.DestroyTarget{Target: id}, effects.CreateToken{Template: effects.RedGoblinToken(), N: 1})(ctx.Game, ctx.Item)
		},
	})
}
