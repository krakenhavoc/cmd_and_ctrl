package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana's Defeat — {B} Sorcery (Hour of Devastation):
//
//	"Destroy target black creature or black planeswalker. If that permanent
//	 was a Liliana planeswalker, its controller loses 3 life."
//
// The permanent is read as it last existed before it is destroyed (CR 608.2h),
// so its controller and its Liliana subtype are known. The life loss is not
// conditional on the destruction taking place.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c8b38b3-b808-4ad8-b4e5-39e216063ff6",
		Name:         "Liliana's Defeat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target black creature or black planeswalker",
			effects.OfColor("B"), effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			c, ok := effects.OldB238CardTargetLastKnown(ctx)
			if !ok {
				return nil
			}
			wasLiliana := c.IsPlaneswalker() && c.HasSubtype("Liliana")
			if err := effects.DestroyFirstLegalCardTarget(ctx.Game, ctx.Item); err != nil {
				return err
			}
			if !wasLiliana {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), c.Controller, -3)
		},
	})
}
