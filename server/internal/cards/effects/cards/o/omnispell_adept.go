package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omnispell Adept — Creature — Human Wizard {4}{U}, 3/4 (Guilds of Ravnica,
// #5525):
//
//	"{2}{U}, {T}: You may cast an instant or sorcery spell from your hand
//	 without paying its mana cost."
//
// Surtland Elementalist's free cast (OfferFreeHandCast, keyed so a table waiting
// on it is a restore point): as the ability resolves its controller may pick one
// instant or sorcery card in their hand and cast it for {0}, at any timing
// (CR 608.2g); declining casts nothing. The {T} cost needs the Adept untapped and
// without summoning sickness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b575449b-591d-4357-92b7-55137fc7966a",
		Name:         "Omnispell Adept",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: You may cast an instant or sorcery spell from your hand without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				effects.OfferFreeHandCast(effects.NewContext(g, item), item.Controller, omnispellAdeptFreeCast,
					"an instant or sorcery spell", effects.FreeInstantOrSorcery())
				return nil
			},
		}},
	})
}

// omnispellAdeptFreeCast is the offer's keyed continuation (#4246). An on-disk
// key: never renamed, never reused.
var omnispellAdeptFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/omnispell-adept-cast", "Omnispell Adept")
