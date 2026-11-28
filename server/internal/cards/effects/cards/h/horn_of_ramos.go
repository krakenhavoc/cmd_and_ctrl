package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horn of Ramos — Artifact {3} (Mercadian Masques, S82 OLD batch 699, #5619):
//
//	"{T}: Add {G}.
//	 Sacrifice this artifact: Add {G}."
//
// Two mana abilities (S82OldB699RamosAbilities, CR 605.1a): the tap, and the
// sacrifice, which needs no {T} and so works the turn the artifact arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3b9bcf88-7304-48c5-bd46-3e37a93967a5",
		Name:          "Horn of Ramos",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: effects.S82OldB699RamosAbilities("{G}"),
	})
}
