package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Augmenting Automaton — Artifact Creature — Construct {1}, 1/1 (Aether Revolt,
// #5536):
//
//	"{1}{B}: This creature gets +1/+1 until end of turn."
//
// Tattered Apparition's self-pump row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ec0e58c-99b5-4ebf-8053-f59c30b86ef7",
		Name:         "Augmenting Automaton",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Augmenting Automaton", "{1}{B}", 1, 1)},
	})
}
