package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pharika's Disciple — Creature — Centaur Warrior {3}{G}, 2/3 (Magic Origins, #5549):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
//	 Renown 1 (When this creature deals combat damage to a player, if it isn't renowned, put a +1/+1 counter on it and it becomes renowned.)"
//
// Keywords only: deathtouch, and renown, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7798312-657c-4f4a-9e2d-906284caf50f",
		Name:            "Pharika's Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "renown 1"},
	})
}
