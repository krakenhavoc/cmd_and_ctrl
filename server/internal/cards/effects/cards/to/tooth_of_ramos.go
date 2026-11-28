package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tooth of Ramos — Artifact {3} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{T}: Add {W}.
//	 Sacrifice this artifact: Add {W}."
//
// Two mana abilities (S82OldB699RamosAbilities, CR 605.1a): the tap, and the
// sacrifice, which needs no {T} and so works the turn the artifact arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "fa4c57b3-6eaa-4938-b41a-0dad3e774d49",
		Name:          "Tooth of Ramos",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.S82OldB699RamosAbilities("{W}"),
	})
}
