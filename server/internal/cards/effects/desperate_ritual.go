package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desperate Ritual — Instant — Arcane {1}{R}:
//
//	"Add {R}{R}{R}.
//	Splice onto Arcane {1}{R}"
//
// Spliced, the spell's controller adds the mana as the spell resolves
// (ADR 0202).
func init() {
	Register(Spec{
		OracleID:     "3620ab8f-b693-4225-bb96-5ca3cb3fa465",
		Name:         "Desperate Ritual",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{R}"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
