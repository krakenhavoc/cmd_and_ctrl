package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skull of Ramos — Artifact {3} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{T}: Add {B}.
//	 Sacrifice this artifact: Add {B}."
//
// Two mana abilities (S82OldB699RamosAbilities, CR 605.1a): the tap, and the
// sacrifice, which needs no {T} and so works the turn the artifact arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b5ae2532-e642-47d1-bb5f-53f408e2fdc2",
		Name:          "Skull of Ramos",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.S82OldB699RamosAbilities("{B}"),
	})
}
