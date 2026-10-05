package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyretic Ritual — Instant {1}{R}:
//
//	"Add {R}{R}{R}."
//
// Dark Ritual's shape: a spell, not a mana ability, so it uses the
// stack and can be countered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86ef6474-613f-41fb-931c-d4279b03ed99",
		Name:         "Pyretic Ritual",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
