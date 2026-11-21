package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consulate Skygate — Artifact Creature — Wall {2}, 0/4 (Kaladesh, #5539):
//
//	"Defender
//	 Reach (This creature can block creatures with flying.)"
//
// Printed keywords only; the engine reads each one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "671db0d7-b103-4c13-9914-a444f2ecf9d6",
		Name:            "Consulate Skygate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
	})
}
