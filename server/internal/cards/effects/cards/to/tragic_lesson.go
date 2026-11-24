package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tragic Lesson — Instant {2}{U} (Jumpstart 2022, #5948):
//
//	"Draw two cards. Then discard a card unless you return a land you
//	 control to its owner's hand."
//
// After the draw the controller may pick one land they control to return
// (a pick of none is a legal answer); picking none, or controlling no
// land, is the discard.
//
// No simplification.
var tragicLessonReturnOrDiscard = effects.RegisterPermanentsPick("gap/tragic-lesson-return-a-land-or-discard", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		if len(picked.Cards()) == 0 {
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 1})
			return nil
		}
		for _, id := range picked.Cards() {
			if err := (effects.BounceToHand{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	},
})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "516d1ac4-e831-44b6-8266-131dde7fe802",
		Name:         "Tragic Lesson",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			lands := false
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsLand() {
					lands = true
					break
				}
			}
			if !lands {
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: ctx.Controller(), Source: ctx.Source(), N: 1})
				return nil
			}
			return effects.ChoosePermanents{
				Question:   "Tragic Lesson — return a land you control to its owner's hand, or discard a card",
				Candidates: effects.OldB093OptionalLandYouControl,
				Keyed:      tragicLessonReturnOrDiscard,
			}.Apply(ctx)
		},
	})
}
