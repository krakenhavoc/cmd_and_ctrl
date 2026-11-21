package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraul Raider — Creature — Insect Warrior {2}{B}, 2/3 (Guilds of Ravnica, #5525):
//
//	"Menace"
//
// Menace is an engine keyword carried as printed data (CR 702.111b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3c741801-bfa5-4ff6-bf06-8d5edc06388d",
		Name:            "Kraul Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
