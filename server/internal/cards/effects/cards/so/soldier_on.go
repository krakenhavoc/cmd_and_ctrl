package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldier On — Instant {2}{W} (Marvel Super Heroes Commander, #4455):
// "Put two +1/+1 counters on target creature. Untap it." Two steps on
// the one target: the counters, then the untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f522e161-f249-4add-8326-b6ab0de8ba57",
		Name:         "Soldier On",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.OldB346TwoCountersAndUntapTheTarget,
	})
}
