package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unworthy Dead — Creature — Phyrexian Skeleton {1}{B} 1/1 (Urza's Saga, #5624):
//
//	"{B}: Regenerate this creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7454e2cf-5ef2-4623-bac7-c375fd230595",
		Name:         "Unworthy Dead",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
