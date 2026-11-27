package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Swarm — {3}{U} Creature — Faerie (Shadowmoor, S82 older sets):
//
//	"Flying
//	 Faerie Swarm's power and toughness are each equal to the number of blue
//	 permanents you control."
//
// Abominable Treefolk's characteristic-defining count over blue permanents
// (PowerAndToughnessEqualToPermanentsOfColorYouControl), with the printed keyword.
// The creature counts itself, being blue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7178feb-29ce-45d9-a970-2c6003779a00",
		Name:            "Faerie Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.PowerAndToughnessEqualToPermanentsOfColorYouControl("Faerie Swarm", "U", "blue")},
	})
}
