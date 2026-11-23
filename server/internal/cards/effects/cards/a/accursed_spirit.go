package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Accursed Spirit — Creature — Spirit {3}{B}, 3/2 (Magic 2014, #5563):
//
//	"Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)"
//
// Bladetusk Boar's intimidate on a black 3/2: only an artifact creature or a
// black creature can block it (CR 702.13b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e20afd2-507b-4878-a78c-f23b51ec071f",
		Name:            "Accursed Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
	})
}
