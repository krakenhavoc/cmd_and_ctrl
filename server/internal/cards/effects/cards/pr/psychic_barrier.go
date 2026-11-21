package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Barrier — Instant {U}{U} (New Phyrexia, #5575):
//
//	"Counter target creature spell. Its controller loses 1 life."
//
// Countersquall's body over a creature spell, for 1 life: the spell's
// controller is read before it is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e712090-2284-4325-8ee8-38cd0e2bedd8",
		Name:         "Psychic Barrier",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B33CounterTargetThenControllerLosesLife(item, ctx, 1)
		},
	})
}
