package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloud Crusader — Creature — Human Knight {2}{W}{W}, 2/3 (Magic 2011,
// #5579):
//
//	"Flying
//	 First strike"
//
// Both are printed engine keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ef481fa-ff20-426a-af16-fbf761ea5eaf",
		Name:            "Cloud Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
