package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
				c.Power = cmdrB9LandCardsInGraveyard(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Uurg — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}, Sacrifice a land: You gain 2 life.",
			Purpose: game.Purpose{LifeGain: 2},
			Cost:    Plus(ManaCost("{B}{G}"), b08SacrificeALand()),
			Effect:  Do(GainLife{Amount: 2}),
		}},
	})
}

// cmdrB9LandCardsInGraveyard counts the land cards in a player's graveyard.
func cmdrB9LandCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsLand() {
			n++
		}
	}
	return n
}
