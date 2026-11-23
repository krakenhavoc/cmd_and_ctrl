package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subgoyf — Artifact Creature — Lhurgoyf {1}{U}, */1+* (Mystery Booster 2):
//
//	"Subgoyf's power is equal to the number of different subtypes other than
//	 creature types among cards in all graveyards and its toughness is equal to
//	 that number plus 1. (Subtypes include but are absolutely not limited to:
//	 Plains, Equipment, Aura, Arcane, Jace, Curse, Urza's, Ulamog's, etc.)"
//
// A layer 7a characteristic-defining ability (CR 604.3) recounted whenever a
// graveyard changes (Dragonfly Swarm's shape). Every card in every player's
// graveyard is read for its subtypes; the same subtype on several cards counts
// once, and anything that is a creature type (Human, Elf, a Lhurgoyf, ...) is
// left out, wherever it appears. Land, artifact, enchantment, planeswalker and
// spell subtypes all count. A graveyard card's subtypes are those it prints.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e57331c-b5b0-40b7-ba2d-6effa9d0abea",
		Name:         "Subgoyf",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Subgoyf's power is equal to the number of different subtypes other than creature types among cards in all graveyards and its toughness is equal to that number plus 1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := subgoyfDifferentNonCreatureSubtypes(g)
				c.Power, c.Toughness = n, n+1
			},
		}},
	})
}

// subgoyfDifferentNonCreatureSubtypes counts the different subtypes that are
// not creature types among the cards in every graveyard.
func subgoyfDifferentNonCreatureSubtypes(g *game.Game) int {
	seen := map[string]bool{}
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			for _, sub := range c.Effective().Subtypes {
				if !game.IsCreatureType(sub) {
					seen[sub] = true
				}
			}
		}
	}
	return len(seen)
}
