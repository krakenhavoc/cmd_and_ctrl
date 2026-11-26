package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightmare — Creature — Nightmare Horse {5}{B}, */* (Magic Origins, S82 older
// sets):
//
//	"Flying
//	 Nightmare's power and toughness are each equal to the number of Swamps you
//	 control."
//
// A characteristic-defining ability (CR 604.3, layer 7a) counted live as the
// layers are recomputed, like Ashaya, Flora Colossus: every permanent you
// control with the Swamp subtype, a land that has been made a Swamp included
// (CR 205.3i). The type is a word of the source's text (CR 612.2). It is 0/0 and
// dies with no Swamp.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "375932e6-1b3e-48dc-8154-9b664c3add34",
		Name:            "Nightmare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Nightmare's power and toughness are each equal to the number of Swamps you control.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				swamp := source.TextSubtype("Swamp")
				n := effects.CountControlled(g, source.Controller, func(p game.Card) bool { return p.HasSubtype(swamp) })
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
