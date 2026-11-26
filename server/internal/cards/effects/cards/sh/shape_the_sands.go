package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shape the Sands — Instant {G} (Dragons of Tarkir, #5551):
//
//	"Target creature gets +0/+5 and gains reach until end of turn. (It can block creatures with flying.)"
//
// Rush of Adrenaline's pump-and-keyword body (+0/+5 and reach).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "d96506b1-845d-4a9b-8414-456bb457d80f",
		Name:         "Shape the Sands",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 5, "reach")),
		OnResolve:    effects.SocB27TargetGetsAndGains(0, 5, "reach", "Shape the Sands"),
	})
}
