package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Insatiable Harpy — Creature — Harpy {2}{B}{B}, 2/2 (Theros, #5561):
//
//	"Flying, lifelink"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6774c0ce-103e-471b-8de0-f325805674ca",
		Name:            "Insatiable Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
