package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kagemaro, First to Suffer — Legendary Creature — Demon Spirit {3}{B}{B},
// */* (Saviors of Kamigawa, #5601):
//
//	"Kagemaro's power and toughness are each equal to the number of cards
//	 in your hand.
//	 {B}, Sacrifice Kagemaro: All creatures get -X/-X until end of turn,
//	 where X is the number of cards in your hand."
//
// The first line is a characteristic-defining ability (CR 604.3, layer 7a),
// read live. X of the second is counted as the ability resolves, after the
// sacrifice cost, so a hand that changed in response is the hand counted
// (CR 608.2h). Only the creatures on the battlefield then are affected
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da3e399f-6801-4e4d-9b1d-ba15fcfb7ee4",
		Name:         "Kagemaro, First to Suffer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Kagemaro's power and toughness are each equal to the number of cards in your hand.",
			Layer: game.Layer7PT,
			// A hand move drops the layer cache while Kagemaro is out.
			DependsOnHandSize: true,
			SubLayer:          game.SubLayer7A_CDA,
			AppliesTo:         effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CardsInHand(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice Kagemaro: All creatures get -X/-X until end of turn, where X is the number of cards in your hand.",
			Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true}},
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.SokB1CardsInYourHand(ctx)
				if x <= 0 {
					return nil
				}
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -x, Toughness: -x, Label: "Kagemaro — -X/-X"}.Apply(ctx)
			},
		}},
	})
}
