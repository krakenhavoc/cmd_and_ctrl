package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maraxus of Keld — Legendary Creature — Human Warrior {4}{R}{R}, */* (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Maraxus's power and toughness are each equal to the number of untapped artifacts, creatures, and
//	 lands you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a), counted live as the layers are recomputed: each untapped
// permanent of those types counts once, Maraxus itself included while it is untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af4b67c5-53dc-4e4d-9ec4-fa51cc3cd49a",
		Name:         "Maraxus of Keld",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Maraxus's power and toughness are each equal to the number of untapped artifacts, creatures, and lands you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB779UntappedPermanentsCount(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
