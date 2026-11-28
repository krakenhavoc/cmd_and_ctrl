package cardshu

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Iron Tusk Elephant — Creature — Elephant {4}{W}, 3/3 (Mirage, S82 OLD b807):
//
//	"Trample"
//
// Trample is an engine keyword carried as printed data (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a299470a-72ea-4848-833b-dc894db7be15",
		Name:            "Iron Tusk Elephant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
