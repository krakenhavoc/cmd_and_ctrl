package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lash of Thorns — Instant {B} (Throne of Eldraine, #5516):
//
//	"Target creature gets +2/+1 and gains deathtouch until end of turn."
//
// Stonewood Invocation's pump-and-keyword body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "642d3494-7a6a-4aec-ba64-7c0c467ada99",
		Name:         "Lash of Thorns",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1, "deathtouch")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.TspB1PumpAndGrantUntilEOT(2, 1, "deathtouch", "Lash of Thorns"),
	})
}
