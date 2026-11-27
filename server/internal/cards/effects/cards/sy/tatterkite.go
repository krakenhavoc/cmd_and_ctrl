package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tatterkite — Artifact Creature — Scarecrow {3}, 2/1 (Shadowmoor, #5591):
//
//	"Flying
//	 This creature can't have counters put on it."
//
// Flying is a printed keyword. The counter bar is Melira's Keepers' layer 6
// static on itself, which the engine reads wherever a counter would be
// placed, paid or chosen (CR 101.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9505f124-f152-48ce-b0d3-097abdda8e3e",
		Name:            "Tatterkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature can't have counters put on it.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CantHaveCounters = true
			},
		}},
	})
}
