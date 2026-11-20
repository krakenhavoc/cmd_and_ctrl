package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dfa8415d-2474-46ca-b92f-6e12e3268cf5",
		Name:         "Unexpected Fangs",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    ResolveSteps(ikoCounterOnTheTarget(game.CounterPlusOne), ikoCounterOnTheTarget(game.CounterLifelink)),
	})
}
