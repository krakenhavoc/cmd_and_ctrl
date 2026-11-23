package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-UK — Legendary Creature — Spider Human Hero {3}{W}, 3/4 (Marvel
// Super Heroes Commander, #4455):
//
//	"Web-slinging {2}{W} (You may cast this spell for {2}{W} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 At the beginning of your end step, if two or more creatures entered
//	 the battlefield under your control this turn, you draw a card and
//	 gain 2 life."
//
// Web-slinging is Spider-Sense's alternative cost (mana plus a returned
// tapped creature, paid at announce). The end-step trigger is an
// intervening if (CR 603.4) read off the turn's entry tally for creatures
// (Iron Man, Master of Machines' artifacts, Avengers Assemble!'s end
// step): tokens count, and so does Spider-UK himself, as he entered under
// your control this turn. The condition is checked as the end step begins
// and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebf8d006-2997-431b-b4cb-9692741b45ca",
		Name:         "Spider-UK",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {2}{W}",
			ManaCost:     "{2}{W}",
			ReturnToHand: effects.PermanentYouControl("a tapped creature you control", effects.Creature(), effects.B751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(spiderUKCond, effects.On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
						g.EnteredWithCardTypeThisTurn(source.Controller, "creature") >= 2
				},
				"Spider-UK — draw a card and gain 2 life", effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: 2}))),
		},
	})
}

// spiderUKCond is Spider-UK's intervening "if", re-read as the trigger
// resolves.
var spiderUKCond = effects.CondBoard{
	Name: "two or more creatures entered the battlefield under your control this turn",
	Test: func(ctx *effects.Context) bool {
		return ctx.Game.EnteredWithCardTypeThisTurn(ctx.Controller(), "creature") >= 2
	},
}
