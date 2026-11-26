package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tread Upon — Instant {1}{G} (Dragons of Tarkir, #5551):
//
//	"Target creature gets +2/+2 and gains trample until end of turn."
//
// Rush of Adrenaline's pump-and-keyword body (+2/+2 and trample).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b3bdbc6-cde7-4187-b13c-82fd02ae4037",
		Name:         "Tread Upon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		OnResolve:    effects.SocB27TargetGetsAndGains(2, 2, "trample", "Tread Upon"),
	})
}
