package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lash of the Whip — Instant {4}{B} (Theros, #5561):
//
//	"Target creature gets -4/-4 until end of turn."
//
// The shared until-end-of-turn shrink of the one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2ac369a-f6d4-4b58-8695-a7c5b906e999",
		Name:         "Lash of the Whip",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, -4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(-4, -4, "Lash of the Whip — -4/-4 until end of turn"),
	})
}
