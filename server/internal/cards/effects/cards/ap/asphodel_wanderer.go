package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Asphodel Wanderer — Creature — Skeleton Soldier {B}, 1/1 (Theros, #5561):
//
//	"{2}{B}: Regenerate this creature."
//
// The shared regeneration row (regenerate_this.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6ffcbc9-0990-4d95-a995-1b2f321256d7",
		Name:         "Asphodel Wanderer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{2}{B}")},
	})
}
