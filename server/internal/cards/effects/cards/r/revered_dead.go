package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revered Dead — Creature — Spirit Soldier {1}{W}, 1/1 (Planar Chaos, #5595):
//
//	"{W}: Regenerate this creature."
//
// Regeneration on the ability's own source (CR 701.19), lrwB1RegenerateThisRow.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96a57ce0-f1ab-4af7-8c18-8e35aad11780",
		Name:         "Revered Dead",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{W}")},
	})
}
