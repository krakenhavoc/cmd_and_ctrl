package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vile Aggregate — Creature — Eldrazi Drone {2}{R}, */5 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Vile Aggregate's power is equal to the number of colorless creatures you
//	 control.
//	 Trample
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)"
//
// A characteristic-defining ability (CR 604.3, CR 208.2a) in layer 7a, counting
// the colorless creatures its controller controls through the layers, itself
// included (it is colorless by devoid).
//
// No simplification.
func init() {
	const name = "Vile Aggregate"
	effects.Register(effects.Spec{
		OracleID:        "4a6b254e-76fb-4712-ac34-5e26c311b2e4",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
		Triggered:       []game.TriggeredAbility{effects.OldB306Ingest(name)},
		Static: []game.StaticAbility{{
			Label:     "Vile Aggregate's power is equal to the number of colorless creatures you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, self *game.Card) {
				c.Power = effects.B16ColorlessCreaturesControlled(g, self.Controller)
			},
		}},
	})
}
