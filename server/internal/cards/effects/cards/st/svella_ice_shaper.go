package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Svella, Ice Shaper — Legendary Snow Creature — Troll Warrior {1}{R}{G}, 2/4:
//
//	"{3}, {T}: Create a colorless snow artifact token named Icy Manalith with
//	 '{T}: Add one mana of any color.'
//	 {6}{R}{G}, {T}: Look at the top four cards of your library. You may cast a
//	 spell from among them without paying its mana cost. Put the rest on the
//	 bottom of your library in a random order."
//
// The first ability makes the Icy Manalith catalog token (a snow artifact, so its
// mana is snow mana). The second is Aetherworks Marvel's look-and-cast over the
// top four: a private look, a free cast of a nonland pick, the rest to the
// bottom in a random order. Both tap Svella, so neither can be activated the
// turn she arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7fd8a508-335c-4d5a-9916-23bf26dfff38",
		Name:         "Svella, Ice Shaper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}: Create a colorless snow artifact token named Icy Manalith with \"{T}: Add one mana of any color.\"",
				Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.OldB071IcyManalithToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{6}{R}{G}, {T}: Look at the top four cards of your library. You may cast a spell from among them without paying its mana cost. Put the rest on the bottom of your library in a random order.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{6}{R}{G}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.LookAtTopAndOfferFreeCast(g, ctx.Controller(), ctx.Source(), svellaFreeCast, 4, effects.AnySpell)
				},
			},
		},
	})
}

// svellaFreeCast is the look-and-cast pick's keyed continuation, so a table
// waiting on it is a restore point. An on-disk key: never renamed, never reused.
var svellaFreeCast = effects.RegisterFreeLibraryCastOffer("choose-cards/svella-ice-shaper-cast", "Svella, Ice Shaper")
