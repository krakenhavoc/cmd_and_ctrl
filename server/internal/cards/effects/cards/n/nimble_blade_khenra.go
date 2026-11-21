package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nimble-Blade Khenra — Creature — Jackal Warrior {1}{R}, 1/3
// (Amonkhet, #5535):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)"
//
// Prowess is a printed keyword the engine runs (see Lightshell Duo).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "09c68fb5-2000-4bb8-b03b-94b093e50bca",
		Name:            "Nimble-Blade Khenra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
