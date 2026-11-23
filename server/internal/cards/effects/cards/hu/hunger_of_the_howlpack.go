package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunger of the Howlpack — Instant {G} (Jumpstart 2022, #5948):
//
//	"Put a +1/+1 counter on target creature.
//	 Morbid — Put three +1/+1 counters on that creature instead if a creature
//	 died this turn."
//
// Brimstone Volley's morbid read (any creature, any controller) as the spell
// resolves, choosing between one counter and three on the one target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "296dee6b-6ec1-4ea4-8dc3-23080dd37646",
		Name:         "Hunger of the Howlpack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.GapB170HungerCounters,
	})
}
