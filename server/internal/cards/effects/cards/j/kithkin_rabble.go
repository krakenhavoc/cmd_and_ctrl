package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kithkin Rabble — {3}{W} Creature — Kithkin (Shadowmoor, S82 older sets):
//
//	"Vigilance
//	 Kithkin Rabble's power and toughness are each equal to the number of white
//	 permanents you control."
//
// Abominable Treefolk's characteristic-defining count over white permanents
// (PowerAndToughnessEqualToPermanentsOfColorYouControl), with the printed keyword.
// The creature counts itself, being white.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17386f88-dbf0-47ae-91d1-6b1aaceee05f",
		Name:            "Kithkin Rabble",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{effects.PowerAndToughnessEqualToPermanentsOfColorYouControl("Kithkin Rabble", "W", "white")},
	})
}
