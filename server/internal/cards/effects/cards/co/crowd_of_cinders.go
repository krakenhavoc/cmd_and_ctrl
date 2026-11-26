package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crowd of Cinders — {3}{B} Creature — Elemental (Shadowmoor, S82 older sets):
//
//	"Fear
//	 Crowd of Cinders's power and toughness are each equal to the number of black
//	 permanents you control."
//
// Abominable Treefolk's characteristic-defining count over black permanents
// (PowerAndToughnessEqualToPermanentsOfColorYouControl), with the printed keyword.
// The creature counts itself, being black.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc071117-08aa-47c3-ae4a-4b6c630befa3",
		Name:            "Crowd of Cinders",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Static:          []game.StaticAbility{effects.PowerAndToughnessEqualToPermanentsOfColorYouControl("Crowd of Cinders", "B", "black")},
	})
}
