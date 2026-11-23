package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyretic Ritual — Instant {1}{R}:
//
//	"Add {R}{R}{R}."
//
// Dark Ritual's shape: a spell, not a mana ability, so it uses the
// stack and can be countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86ef6474-613f-41fb-931c-d4279b03ed99",
		Name:         "Pyretic Ritual",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
