package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraul Stinger — Creature — Insect Assassin {2}{G}, 2/2 (War of the Spark, #5520):
//
//	"Deathtouch"
//
// The keywords are engine keywords carried as printed data; combat does
// the rest.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8481d91e-0242-4edc-8f4a-ffa47cf2e097",
		Name:            "Kraul Stinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
