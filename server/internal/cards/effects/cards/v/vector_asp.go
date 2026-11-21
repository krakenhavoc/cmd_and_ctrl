package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vector Asp — Artifact Creature — Phyrexian Snake {1}, 1/1 (Scars of
// Mirrodin, #5577):
//
//	"{B}: This creature gains infect until end of turn."
//
// A self grant until end of turn (somB1SelfGainsUntilEOT); infect is the
// engine's damage tail (CR 702.90), read as damage is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1950067a-cc57-42a8-b1ee-f03e1b50591b",
		Name:         "Vector Asp",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SomB1SelfGainsUntilEOT("Vector Asp", "{B}", "infect")},
	})
}
