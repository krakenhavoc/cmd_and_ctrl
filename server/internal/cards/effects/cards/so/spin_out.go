package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spin Out — Instant {1}{B}{B} (Aetherdrift, #4464):
//
//	"Destroy target creature or Vehicle."
//
// Broken Wings' destroy (destroyTheTargetPermanent) with Bounce Off's
// "creature or Vehicle" target, so an uncrewed Vehicle is a legal
// target. One target: if it has gone by resolution the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1b6795a-1ac6-4edb-bcee-4db1f1fbe8ea",
		Name:         "Spin Out",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
