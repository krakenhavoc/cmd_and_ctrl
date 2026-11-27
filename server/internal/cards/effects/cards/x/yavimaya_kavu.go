package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yavimaya Kavu — Creature — Kavu {2}{R}{G}, */* (Invasion, S82 older sets batch 664, #5616):
//
//	"Yavimaya Kavu's power is equal to the number of red creatures on the
//	 battlefield.
//	 Yavimaya Kavu's toughness is equal to the number of green creatures on
//	 the battlefield."
//
// Beast of Burden's layer 7a characteristic-defining ability (CR 604.3, 613.4a)
// with two counts: red creatures for power and green creatures for toughness,
// every creature on the battlefield whoever controls it, the Kavu itself (red
// and green) included in both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "285c9f49-4aeb-468c-bb41-f5ccb5257e2b",
		Name:         "Yavimaya Kavu",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				c.Power = effects.OldB664CreaturesOfColorOnTheBattlefield(g, "R")
				c.Toughness = effects.OldB664CreaturesOfColorOnTheBattlefield(g, "G")
			},
		}},
	})
}
