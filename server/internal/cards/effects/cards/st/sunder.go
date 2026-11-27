package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunder — Instant {3}{U}{U} (Urza's Saga, S82 OLD batch 729, #5624):
//
//	"Return all lands to their owners' hands."
//
// Every land on the battlefield, whoever controls it, returns as one
// simultaneous event (Evacuation's bounce over lands). A land creature is a
// land and returns too; tokens cease to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fddc93a2-912e-402a-9d0e-34f778e659ea",
		Name:         "Sunder",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.Land()}.Apply(ctx)
		},
	})
}
