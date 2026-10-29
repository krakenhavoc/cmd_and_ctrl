package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// powerIsCardsInYourHand is the characteristic-defining ability "this
// creature's power is equal to the number of cards in your hand"
// (CR 604.3, layer 7a): Robobrain War Mind and Minas Tirith Garrison. It
// sets power only, so a card with a printed toughness keeps it.
// DependsOnHandSize drops the layer cache whenever a hand changes while the
// permanent is on the battlefield.
func powerIsCardsInYourHand() game.StaticAbility {
	return game.StaticAbility{
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Power = n
		},
	}
}
