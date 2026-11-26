package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Umbra Stalker — {4}{B}{B}{B} Creature — Elemental, */*:
//
//	"Chroma — Umbra Stalker's power and toughness are each equal to the number
//	 of black mana symbols in the mana costs of cards in your graveyard."
//
// A characteristic-defining ability (CR 604.3, CR 208.3): layer 7a.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51afd09d-7eee-41c9-bc42-f3f61a176479",
		Name:         "Umbra Stalker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Chroma — power and toughness are each equal to the number of black mana symbols in the mana costs of cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB521ChromaOfGraveyard(g, source.Controller, source.TextColor("B"))
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
