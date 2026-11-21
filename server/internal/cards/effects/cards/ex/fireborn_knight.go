package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fireborn Knight — Creature — Human Knight {R/W}{R/W}{R/W}{R/W}, 2/3 (Throne of Eldraine, #5516):
//
//	"Double strike
//	 {R/W}{R/W}{R/W}{R/W}: This creature gets +1/+1 until end of turn."
//
// The shared pump row (thisGetsRow) at a hybrid cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b77695da-86fb-4675-b2d3-51732ad0325f",
		Name:            "Fireborn Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{R/W}{R/W}{R/W}{R/W}", 1, 1, "Fireborn Knight")},
	})
}
