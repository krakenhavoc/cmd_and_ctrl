package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vulpine Goliath — Creature — Fox {4}{G}{G}, 6/5 (Theros, #5561):
//
//	"Trample"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4c96980-31ed-494e-89db-c0a0f1fd23fd",
		Name:            "Vulpine Goliath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
