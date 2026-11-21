package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Mist — Creature — Illusion {1}{U}, */*:
//
//	"Flying
//	 Krovikan Mist's power and toughness are each equal to the number of
//	 Illusions on the battlefield."
//
// A characteristic-defining ability (CR 604.3, layer 7a) counted live over
// every Illusion on the battlefield, any controller's, itself included
// (Beanstalk Giant's shape over a different count).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cae42b6-8b09-4188-8e79-203eebc77585",
		Name:            "Krovikan Mist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Krovikan Mist's power and toughness are each equal to the number of Illusions on the battlefield.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				illusion := source.TextSubtype("Illusion")
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.HasSubtype(illusion) {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
