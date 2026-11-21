package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorn Captain — Creature — Gorn Soldier {3}{G}, 4/3 (Star Trek Commander, #4974):
//
//	"Trample (This creature can deal excess combat damage to the player it's attacking.)"
//
// The keyword is the whole card: PrintedKeywords feeds the synthesized
// layer-6 static, and the engine's combat rules read it from there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60ff591d-170d-410a-9710-510e811f9827",
		Name:            "Gorn Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
