package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Countersquall — Instant {U}{B} (EDHREC rank 3490):
//
//	"Counter target noncreature spell. Its controller loses 2 life."
//
// Negate with a sting. The spell's controller is read before the
// counter moves it off the stack, and the life is a LOSS, not damage
// — no prevention or damage replacement sees it. A target that left
// the stack in response does nothing at all, the life loss included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6df620b8-1e54-4f63-8556-12c75e5679af",
		Name:         "Countersquall",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B33CounterTargetThenControllerLosesLife(item, ctx, 2)
		},
	})
}
