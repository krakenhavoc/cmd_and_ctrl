package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monstrous War-Leech — Creature — Leech Horror {3}{B}, */* (Dominaria
// United):
//
//	"Kicker {U}
//	 As this creature enters, if it was kicked, mill four cards.
//	 Monstrous War-Leech's power and toughness are each equal to the greatest
//	 mana value among cards in your graveyard."
//
// Urborg Lhurgoyf's shape: the mill is an as-enters clause (CR 614.12), part
// of the entry and off the stack, so nothing responds to it, and it reads the
// kicker record the cast carries onto the permanent. The size is a layer 7a
// characteristic-defining ability reading the controller's graveyard live,
// so the milled cards count the moment it lands; an empty graveyard is 0/0
// and the creature dies as a state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b92f2a4c-36dc-40a6-a874-7405fc97fb7a",
		Name:          "Monstrous War-Leech",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{U}")},
		AsEnters: func(self *game.Card, ctx *effects.Context) error {
			if game.CardKickedTimes(*self) == 0 {
				return nil
			}
			return effects.MillCards{Player: self.Controller, N: 4}.Apply(ctx)
		},
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the greatest mana value among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.RfCreatureCGreatestManaValueInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
