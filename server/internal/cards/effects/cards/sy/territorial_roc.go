package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Roc — Creature — Bird {1}{W}, 1/3 (Dragons of Tarkir,
// #5551):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data (CR 702.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "adb5883d-44cc-455e-94dc-20e1c1f97d68",
		Name:            "Territorial Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
