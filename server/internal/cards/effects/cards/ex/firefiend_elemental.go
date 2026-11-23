package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firefiend Elemental — Creature — Elemental {3}{R}, 3/2 (Magic Origins, #5549):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your control.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: haste, and renown, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "59c898d8-14f0-4645-84e7-aeedb2ee6c5e",
		Name:            "Firefiend Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "renown 1"},
	})
}
