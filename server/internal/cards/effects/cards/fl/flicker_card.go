package cardsfl

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flicker — Sorcery {1}{W} (Urza's Destiny, S82 #5621):
//
//	"Exile target nontoken permanent, then return it to the battlefield under its
//	 owner's control."
//
// Acrobatic Maneuver's blink (the Flicker primitive) over any nontoken permanent:
// it returns under its owner's control as a new object (CR 400.7), with no
// counters and no Auras, and re-triggers its enters abilities. A target that is
// illegal on resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19ea326f-45ff-4bc0-a957-7d3f22dd8c6f",
		Name:         "Flicker",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nontoken permanent", effects.Not(effects.IsTokenPredicate())),
		OnResolve:    effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.Flicker{Target: id}.Apply(ctx) })),
	})
}
