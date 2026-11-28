package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aysen Crusader — Creature — Human Knight {2}{W}{W}, 2+*/2+* (Homelands, S82
// OLD b826, #5635):
//
//	"Aysen Crusader's power and toughness are each equal to 2 plus the
//	 number of Soldiers and Warriors you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a), read live: each
// permanent its controller controls that is a Soldier, a Warrior or both
// counts once. The Crusader is a Human Knight, so it does not count itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ebbe05c-7cc4-4845-87e5-ab93d7ac9623",
		Name:         "Aysen Crusader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Aysen Crusader's power and toughness are each equal to 2 plus the number of Soldiers and Warriors you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 2 + effects.OldB826CountSubtypes(g, source.Controller, "Soldier", "Warrior")
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
