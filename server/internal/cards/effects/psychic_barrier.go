package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psychic Barrier — Instant {U}{U} (New Phyrexia, #5575):
//
//	"Counter target creature spell. Its controller loses 1 life."
//
// Countersquall's body over a creature spell, for 1 life: the spell's
// controller is read before it is countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e712090-2284-4325-8ee8-38cd0e2bedd8",
		Name:         "Psychic Barrier",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b33CounterTargetThenControllerLosesLife(item, ctx, 1)
		},
	})
}
