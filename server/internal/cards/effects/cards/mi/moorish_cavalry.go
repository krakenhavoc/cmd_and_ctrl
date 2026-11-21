package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moorish Cavalry — Creature — Human Knight {2}{W}{W}, 3/3 (Arabian
// Nights, #5641):
//
//	"Trample"
//
// Trample is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bccdc42f-9e55-4f8e-a29e-ef39f38add6a",
		Name:            "Moorish Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
