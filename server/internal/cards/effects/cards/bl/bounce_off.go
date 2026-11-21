package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bounce Off — Instant {U} (Aetherdrift, #4464):
//
//	"Return target creature or Vehicle to its owner's hand."
//
// Void Squall's bounce (bounceTheTarget) with the target a creature or
// a Vehicle, crewed or not (Back on Track's "creature or Vehicle"
// predicate, on the battlefield). One
// target, so a target that has gone by resolution counters the spell
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9c29e6e-5949-4599-bf74-b9539e753f76",
		Name:         "Bounce Off",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		OnResolve:    effects.BounceTheTarget,
	})
}
