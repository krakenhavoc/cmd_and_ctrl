package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightwing Shade — Creature — Shade {4}{B}, 2/2 (Magic 2011, #5579):
//
//	"Flying
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Flying is the printed keyword. The pump is Thornling's shared
// thisGetsRow, a layer 7c change that ends at cleanup and stacks with
// each activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e679ef01-8d12-4246-850a-efed7f4316b2",
		Name:            "Nightwing Shade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{1}{B}", 1, 1, "Nightwing Shade")},
	})
}
