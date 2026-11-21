package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Blademaster — Creature — Vedalken Soldier {2}{U}, 2/3 (Kaladesh, #5539):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Printed keywords only; the engine reads each one (prowess is its
// own trigger, CR 702.108a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a8405f5-ef37-460e-8e0b-55378f3b407b",
		Name:            "Vedalken Blademaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
