package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unexpected Fangs — Instant {1}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Put a +1/+1 counter and a lifelink counter on target creature."
//
// Both counters on the target if it is still legal as the spell
// resolves (CR 608.2b); the lifelink counter gives it lifelink for as
// long as it stays (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfa8415d-2474-46ca-b92f-6e12e3268cf5",
		Name:         "Unexpected Fangs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.IkoCounterOnTheTarget(game.CounterPlusOne), effects.IkoCounterOnTheTarget(game.CounterLifelink)),
	})
}
