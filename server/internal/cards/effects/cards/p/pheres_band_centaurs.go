package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pheres-Band Centaurs — Creature — Centaur Warrior {4}{G}, 3/7 (Theros, #5561):
//
// A vanilla creature: no rules text, so the spec is its identity.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4aebea34-a05b-4626-97bc-f15fa177ef05",
		Name:         "Pheres-Band Centaurs",
		Completeness: effects.CompletenessFull,
	})
}
