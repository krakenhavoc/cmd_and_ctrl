package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungal Infection — Instant {B} (Jumpstart 2022, #5948):
//
//	"Target creature gets -1/-1 until end of turn. Create a 1/1 green
//	 Saproling creature token."
//
// A -1/-1 on the target and a Saproling for the caster, in that order. With
// its one target illegal as it resolves the spell does nothing at all, so
// no Saproling is made (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04a8858e-9033-403e-9346-5a2ad1bdf680",
		Name:         "Fungal Infection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		OnResolve:    effects.GapB170MinusOneAndSaproling,
	})
}
