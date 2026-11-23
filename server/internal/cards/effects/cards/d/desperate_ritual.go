package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Ritual — Instant — Arcane {1}{R}:
//
//	"Add {R}{R}{R}.
//	Splice onto Arcane {1}{R}"
//
// Spliced, the spell's controller adds the mana as the spell resolves
// (ADR 0202).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3620ab8f-b693-4225-bb96-5ca3cb3fa465",
		Name:         "Desperate Ritual",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{R}"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
