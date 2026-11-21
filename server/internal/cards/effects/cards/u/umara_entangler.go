package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Umara Entangler — Creature — Merfolk Rogue Ally {1}{U}, 2/1 (Oath of the Gatewatch, #5545):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Prowess is the engine keyword Bloodfire Expert uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d32cf19d-a974-4161-851d-dcbf37069e26",
		Name:            "Umara Entangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
