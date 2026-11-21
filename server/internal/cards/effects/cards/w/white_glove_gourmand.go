package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// White Glove Gourmand — Creature — Human Noble {2}{W}{B}, 2/2 (Fallout,
// #5948):
//
//	"When this creature enters, create two 1/1 white Human Soldier
//	 creature tokens.
//	 At the beginning of your end step, if another Human died under your
//	 control this turn, create a Food token."
//
// The end-step trigger is an intervening "if" (CR 603.4), true as the step
// begins and again as the ability resolves. It reads this turn's zone
// moves: a creature that went from the battlefield to a graveyard
// under the Gourmand's controller's control (the controller it had as it
// died, CR 603.10a) and was a Human as it last existed (a changeling counts),
// other than the Gourmand itself. A Human token counts, so does one that
// has since left the graveyard.
//
// No simplification.
func init() {
	cond := effects.CondBoard{
		Name: "another Human died under your control this turn",
		Test: func(ctx *effects.Context) bool {
			src, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
			return ok && effects.AnotherHumanDiedUnderYourControlThisTurn(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID, src.TextSubtype("Human"))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "3771c0d7-cea3-4c03-933d-f236f32ef2db",
		Name:         "White Glove Gourmand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("White Glove Gourmand — create two 1/1 white Human Soldier creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 2})),
			effects.InterveningIf(cond, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && effects.AnotherHumanDiedUnderYourControlThisTurn(g, source.Controller, source.InstanceID, source.TextSubtype("Human"))
			}, "White Glove Gourmand — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}))),
		},
	})
}
