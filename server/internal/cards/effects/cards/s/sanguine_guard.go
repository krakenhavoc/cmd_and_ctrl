package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Guard — Creature — Phyrexian Zombie Knight {1}{B}{B} 2/2 (Urza's
// Saga, #5624):
//
//	"First strike
//	 {1}{B}: Regenerate this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d86cc093-5db9-437a-a5bd-77d6faf97ca0",
		Name:            "Sanguine Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
