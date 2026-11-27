package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primeval Shambler — Creature — Horror Mercenary {4}{B} 3/3 (Eighth Edition,
// S82 older sets batch 619, #5607):
//
//	"{B}: This creature gets +1/+1 until end of turn."
//
// Balshan Collaborator's firebreathing row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a2722c6-04a0-4790-be45-dafb7832bb41",
		Name:         "Primeval Shambler",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Primeval Shambler", "{B}", 1, 1)},
	})
}
