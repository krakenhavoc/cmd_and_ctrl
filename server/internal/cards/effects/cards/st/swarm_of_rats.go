package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarm of Rats — Creature — Rat {1}{B}, */1 (Eighth Edition, #5607):
//
//	"Swarm of Rats's power is equal to the number of Rats you control."
//
// A layer 7a characteristic-defining ability (CR 604.3, 613.4a); the Swarm
// itself is a Rat and counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23720758-f572-45e3-9bec-f6d3b974d7f5",
		Name:         "Swarm of Rats",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Swarm of Rats's power is equal to the number of Rats you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.Controller == source.Controller && p.IsCreature() && p.HasSubtype(source.TextSubtype("Rat")) {
						n++
					}
				}
				c.Power = n
			},
		}},
	})
}
