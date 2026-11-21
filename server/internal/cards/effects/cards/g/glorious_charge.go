package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorious Charge — Instant {1}{W} (Magic 2010, #5585):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// Virtuous Charge's team pump: the creatures you control as it resolves get
// +1/+1 (CR 611.2c), so one that enters later this turn does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "754d5009-7887-4293-b554-b3decddb33f7",
		Name:         "Glorious Charge",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(1, 1, "Glorious Charge — creatures you control get +1/+1"),
	})
}
