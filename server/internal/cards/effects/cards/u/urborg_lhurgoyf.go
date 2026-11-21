package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Lhurgoyf — Creature — Lhurgoyf {1}{G}, */1+*:
//
//	"Kicker {U} and/or {B}
//	 As this creature enters, mill three cards for each time it was
//	 kicked.
//	 Urborg Lhurgoyf's power is equal to the number of creature cards
//	 in your graveyard and its toughness is equal to that number
//	 plus 1."
//
// Two kicker costs (CR 702.33b), declared as printed with Kickers
// (#2153): each is its own once-only toggle, and CardKickedTimes counts
// the two together, which is exactly "each time it was kicked" — 3, 3
// or 6 cards.
//
// The mill is an "as enters" clause (CR 614.12), so it runs from
// AsEnters, off the stack, as the permanent arrives.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "f84d8217-51a4-49ba-9df6-d080ed38f37d",
		Name:          "Urborg Lhurgoyf",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: effects.Kickers("{U}", "{B}"),
		AsEnters: func(self *game.Card, ctx *effects.Context) error {
			return effects.MillCards{Player: self.Controller, N: 3 * game.CardKickedTimes(*self)}.Apply(ctx)
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B11CreatureCardsInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
	})
}
