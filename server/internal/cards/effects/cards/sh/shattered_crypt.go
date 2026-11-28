package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shattered Crypt — Sorcery {X}{B}{B} (Weatherlight, S82 OLD b775):
//
//	"Return X target creature cards from your graveyard to your hand. You lose
//	 X life."
//
// Exactly X targets, X being the announced value (CR 107.3). Every target still
// legal as the spell resolves returns (CR 608.2b); the life loss is X whatever
// came back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77bef290-7229-420a-a5c6-8c1b6fea63df",
		Name:         "Shattered Crypt",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.TargetsCountedByX(effects.TargetCardInGraveyard("X target creature cards from your graveyard",
			effects.YouOwn(), effects.Creature())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			if x := ctx.X(); x > 0 {
				return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -x)
			}
			return nil
		},
	})
}
