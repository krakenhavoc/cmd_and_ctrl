package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splinterfright — Creature — Elemental {2}{G}, */*:
//
//	"Trample
//	 Splinterfright's power and toughness are each equal to the number
//	 of creature cards in your graveyard.
//	 At the beginning of your upkeep, mill two cards."
//
// A layer 7a characteristic-defining ability over the controller's own
// graveyard (Wight of the Reliquary's count). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e54c6fa9-59ca-47dc-9354-91681228371e",
		Name:            "Splinterfright",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B11CreatureCardsInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Splinterfright — mill two cards", effects.Do(effects.MillCards{N: 2})),
		},
	})
}
