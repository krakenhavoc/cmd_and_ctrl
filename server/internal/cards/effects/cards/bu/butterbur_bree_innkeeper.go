package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Butterbur, Bree Innkeeper — Legendary Creature — Human Peasant {2}{G}{W},
// 3/3 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"At the beginning of your end step, if you don't control a Food, create
//	 a Food token."
//
// Pest Rescuer's "if you don't control a …" end-of-turn token: the
// intervening "if" is checked as the step begins and again as the trigger
// resolves (CR 603.4), over every Food artifact the controller has, token
// or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec136f6f-015b-4200-97b6-a2c1bf39e135",
		Name:         "Butterbur, Bree Innkeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you don't control a Food",
				Test: func(ctx *effects.Context) bool { return effects.GapB126YouControlNoFood(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && effects.GapB126YouControlNoFood(g, source.Controller)
				},
				"Butterbur, Bree Innkeeper — create a Food token",
				effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}))),
		},
	})
}
