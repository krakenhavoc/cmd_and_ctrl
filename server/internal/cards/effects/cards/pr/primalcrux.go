package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primalcrux — {G}{G}{G}{G}{G}{G} Creature — Elemental, */*:
//
//	"Trample
//	 Chroma — This creature's power and toughness are each equal to the number
//	 of green mana symbols in the mana costs of permanents you control."
//
// A characteristic-defining ability (CR 604.3): layer 7a. It counts its own
// six symbols.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fa6e2fb-3b92-466a-a2cd-b1e5e9a4bf23",
		Name:            "Primalcrux",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Chroma — power and toughness are each equal to the number of green mana symbols in the mana costs of permanents you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB521ChromaOfPermanents(g, source.Controller, source.TextColor("G"))
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
