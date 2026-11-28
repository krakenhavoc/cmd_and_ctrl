package cardssy

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Talruum Minotaur — Creature — Minotaur Berserker {2}{R}{R}, 3/3 (Mirage,
// S82 OLD b807):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data (CR 702.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5661500-c48f-4f6d-bbe7-c8bd7118d862",
		Name:            "Talruum Minotaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
