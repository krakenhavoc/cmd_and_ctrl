package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Standing Troops — Creature — Human Soldier {2}{W}, 1/4 (Eighth Edition, #5607):
//
//	"Vigilance"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e9d5242-424d-4f61-b625-806295dbb0c7",
		Name:            "Standing Troops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
