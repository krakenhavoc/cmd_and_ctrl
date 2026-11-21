package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grappler Spider — Creature — Spider {1}{G}, 2/1 (Worldwake, #5582):
//
//	"Reach (This creature can block creatures with flying.)"
//
// The printed keywords are engine keywords carried as data, as on
// Giant Spider.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ae0c880-a508-471f-a90b-029c417f6b3b",
		Name:            "Grappler Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
