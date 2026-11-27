package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Troll — Creature — Troll {2}{G}, 2/2 (Eighth Edition, S82 OLD batch 612,
// #5607):
//
//	"{G}: Regenerate this creature."
//
// Drudge Skeletons' row (LrwB1RegenerateThisRow) for {G}: a regeneration shield
// on the Troll itself (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ecc61ff-719a-405d-b02a-01607729d05c",
		Name:         "Horned Troll",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{G}")},
	})
}
