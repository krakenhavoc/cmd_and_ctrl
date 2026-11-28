package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Warthog — Creature — Boar Beast {5}{G}, 5/5 (Judgment):
//
//	"Trample"
//
// Trample is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0eb41c36-5910-41fd-97fc-1cae2332e5c5",
		Name:            "Giant Warthog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
