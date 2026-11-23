package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brambleweft Behemoth — Creature — Elemental {4}{G}{G}, 6/6
// (Hour of Devastation, #5534):
//
//	"Trample"
//
// Every word is an engine keyword carried as printed data.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5070528-cc4c-4910-8047-1ed17efd6496",
		Name:            "Brambleweft Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
