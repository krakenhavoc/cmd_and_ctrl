package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horta — Creature — Elemental {4}{R}{R}, 7/6 (Star Trek, #4973):
//
//	"Trample
//	 Basic landcycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8ac0742-15f5-4cf0-888f-f1c2c97fc144",
		Name:            "Horta",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
