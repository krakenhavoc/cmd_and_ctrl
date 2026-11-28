package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heart of Ramos — Artifact {3} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{T}: Add {R}.
//	 Sacrifice this artifact: Add {R}."
//
// Two mana abilities (S82OldB699RamosAbilities, CR 605.1a): the tap, and the
// sacrifice, which needs no {T} and so works the turn the artifact arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "4c774c6e-c5a0-4018-b494-d3c521d2cac3",
		Name:          "Heart of Ramos",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.S82OldB699RamosAbilities("{R}"),
	})
}
