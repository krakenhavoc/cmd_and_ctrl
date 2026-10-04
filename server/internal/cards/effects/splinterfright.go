package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e54c6fa9-59ca-47dc-9354-91681228371e",
		Name:            "Splinterfright",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b11CreatureCardsInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Splinterfright — mill two cards", Do(MillCards{N: 2})),
		},
	})
}
