package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattered Apparition — Creature — Shade {3}{B}, 2/2 (Dominaria United, S82):
//
//	"Flying
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b38c0575-37f7-413c-9dad-c8d7e6a41b90",
		Name:            "Tattered Apparition",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Tattered Apparition", "{1}{B}", 1, 1)},
	})
}
