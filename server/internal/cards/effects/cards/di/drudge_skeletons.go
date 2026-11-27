package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drudge Skeletons — Creature — Skeleton {1}{B}, 1/1 (Magic 2010, #5585):
//
//	"{B}: Regenerate this creature."
//
// Regeneration is on the ability's own source (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c180ed02-07f4-4538-9bea-8c249234a8e2",
		Name:         "Drudge Skeletons",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{B}")},
	})
}
