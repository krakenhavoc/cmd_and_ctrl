package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalwart Aven — Creature — Bird Soldier {2}{W}, 1/3 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: flying, and renown, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "09cc2cda-14d1-424e-a688-1b49b1b4c392",
		Name:            "Stalwart Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "renown 1"},
	})
}
