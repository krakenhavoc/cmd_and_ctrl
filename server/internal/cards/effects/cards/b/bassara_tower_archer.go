package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bassara Tower Archer — Creature — Human Archer {G}{G}, 2/1 (Journey into Nyx, #5558):
//
//	"Reach, hexproof"
//
// Printed keywords only, which the engine enforces.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39daa265-8d02-4449-9c45-ae8ec2b13e5b",
		Name:            "Bassara Tower Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "hexproof"},
	})
}
