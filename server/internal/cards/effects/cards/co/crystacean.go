package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystacean — Creature — Crab {3}{U}, 1/6 (Foundations Jumpstart):
//
//	"Flash"
//
// Flash is the engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "baddf7ee-f6b9-41e3-b930-57129952cc1d",
		Name:            "Crystacean",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
