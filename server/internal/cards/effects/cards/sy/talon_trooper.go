package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talon Trooper — Creature — Bird Scout {1}{W}{U}, 2/3 (Alara Reborn,
// #5586):
//
//	"Flying"
//
// The printed keyword the engine enforces.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "518d0773-40fa-4b02-a89e-5711c19617cd",
		Name:            "Talon Trooper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
