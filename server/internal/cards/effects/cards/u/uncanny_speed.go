package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uncanny Speed — {1}{R} Instant (Avacyn Restored, #5570):
//
//	"Target creature gets +3/+0 and gains haste until end of turn."
//
// The shared socB27TargetGetsAndGains body at +3/+0 and haste.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85b0a2ea-dbdc-456e-9ce1-faa8c13f84d0",
		Name:         "Uncanny Speed",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "haste")),
		OnResolve:    effects.SocB27TargetGetsAndGains(3, 0, "haste", "Uncanny Speed"),
	})
}
