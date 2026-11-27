package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horde of Boggarts — {3}{R} Creature — Goblin (Shadowmoor, S82 older sets):
//
//	"Menace
//	 Horde of Boggarts's power and toughness are each equal to the number of red
//	 permanents you control."
//
// Abominable Treefolk's characteristic-defining count over red permanents
// (PowerAndToughnessEqualToPermanentsOfColorYouControl), with the printed keyword.
// The creature counts itself, being red.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1229b82f-fc50-4259-83c6-31f706ac2303",
		Name:            "Horde of Boggarts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static:          []game.StaticAbility{effects.PowerAndToughnessEqualToPermanentsOfColorYouControl("Horde of Boggarts", "R", "red")},
	})
}
