package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uurg, Spawn of Turg — Legendary Creature — Frog Beast {B}{B}{G}, */5
// (Foundations Commander):
//
//	"Uurg's power is equal to the number of land cards in your
//	 graveyard.
//	 At the beginning of your upkeep, surveil 1.
//	 {B}{G}, Sacrifice a land: You gain 2 life."
//
// The power is Dragonfly Swarm's layer 7a characteristic-defining
// ability over a land count; the upkeep surveil is the footprinted
// primitive; the life ability pays a land sacrifice as a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0191a93a-f304-4c20-9af4-7cdea82c07cf",
		Name:         "Uurg, Spawn of Turg",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Uurg's power is equal to the number of land cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = b17LandCardsInGraveyard(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Uurg — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}, Sacrifice a land: You gain 2 life.",
			Purpose: game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{B}{G}"), b08SacrificeALand()),
			Effect:  Do(GainLife{Amount: 2}),
		}},
	})
}
