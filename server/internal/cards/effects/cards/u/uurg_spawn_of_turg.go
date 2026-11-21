package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0191a93a-f304-4c20-9af4-7cdea82c07cf",
		Name:         "Uurg, Spawn of Turg",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Uurg's power is equal to the number of land cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = effects.B17LandCardsInGraveyard(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Uurg — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}, Sacrifice a land: You gain 2 life.",
			Purpose: game.Purpose{LifeGain: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{B}{G}"), effects.B08SacrificeALand()),
			Effect:  effects.Do(effects.GainLife{Amount: 2}),
		}},
	})
}
