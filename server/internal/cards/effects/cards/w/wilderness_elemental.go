package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wilderness Elemental — Creature — Elemental {1}{R}{G}, */3 (Commander
// 2016, S82 older sets, #5537):
//
//	"Trample
//	 Wilderness Elemental's power is equal to the number of nonbasic lands
//	 your opponents control."
//
// A characteristic-defining ability (CR 604.3, 613.4a), read live through the
// layered view: a land that stops being basic, or a basic that becomes
// nonbasic, changes the count the moment it does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "478d1acf-93fa-4695-83be-693b839ebf80",
		Name:            "Wilderness Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:    "Wilderness Elemental's power is equal to the number of nonbasic lands your opponents control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, land := range g.BattlefieldCardsForEffect() {
					if land.Controller != source.Controller && land.IsLand() && !effects.IsBasicLand(land) {
						n++
					}
				}
				c.Power = n
			},
		}},
	})
}
