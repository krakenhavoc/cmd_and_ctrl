package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Rider — Creature — Spirit Knight {W}{W}, 2/2 (Innistrad, #5572):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)"
//
// Intimidate is an engine keyword carried as printed data: only an
// artifact creature or a creature that shares a color with it can block
// it (CR 702.13b), Bladetusk Boar's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "713641c4-544e-4c25-ab11-10894be2cb53",
		Name:            "Spectral Rider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
	})
}
