package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleaming Barrier — Artifact Creature — Wall {2}, 0/4 (Foundations,
// #4465):
//
//	"Defender
//	 When this creature dies, create a Treasure token."
//
// Piggy Bank's dies trigger on a wall: defender rides PrintedKeywords,
// and the Treasure is the catalog's Treasure token with its mana
// ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1eb9e401-8975-447b-839d-f7cd23897465",
		Name:            "Gleaming Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Gleaming Barrier — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
