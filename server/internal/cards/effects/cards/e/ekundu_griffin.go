package cardse

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Ekundu Griffin — Creature — Griffin {3}{W}, 2/2 (Mirage, S82 OLD b806):
//
//	"Flying, first strike"
//
// Both are engine keywords carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "852b6e19-4594-41ea-9267-0911df643af5",
		Name:            "Ekundu Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
	})
}
