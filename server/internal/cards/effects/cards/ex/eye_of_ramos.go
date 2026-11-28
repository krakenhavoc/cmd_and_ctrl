package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eye of Ramos — Artifact {3} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{T}: Add {U}.
//	 Sacrifice this artifact: Add {U}."
//
// Two mana abilities (S82OldB699RamosAbilities, CR 605.1a): the tap, and the
// sacrifice, which needs no {T} and so works the turn the artifact arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a1bdea9f-56d0-411a-9da8-601dd6dc6d32",
		Name:          "Eye of Ramos",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.S82OldB699RamosAbilities("{U}"),
	})
}
