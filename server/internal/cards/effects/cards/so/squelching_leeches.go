package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squelching Leeches — Creature — Leech {2}{B}{B}, */* (Journey into Nyx, S82
// OLD, #5558):
//
//	"Squelching Leeches's power and toughness are each equal to the number of
//	 Swamps you control."
//
// Nightmare's characteristic-defining ability (CR 604.3, layer 7a) counted live
// as the layers are recomputed: every permanent you control with the Swamp
// subtype, a land made a Swamp included (CR 205.3i). The type is a word of the
// source's text (CR 612.2). It is 0/0 and dies with no Swamp.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "662a2158-4176-4b4a-be6a-13f4501d22f9",
		Name:         "Squelching Leeches",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Squelching Leeches's power and toughness are each equal to the number of Swamps you control.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				swamp := source.TextSubtype("Swamp")
				n := effects.CountControlled(g, source.Controller, func(p game.Card) bool { return p.HasSubtype(swamp) })
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
