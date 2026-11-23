package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benthic Giant — Creature — Giant {5}{U}, 4/5 (Theros, #5561):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43c449d1-51a9-45e0-8102-823a6811ff54",
		Name:            "Benthic Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
	})
}
