package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fruition — Sorcery {G} (Portal, S82 OLD batch 784, #5630):
//
//	"You gain 1 life for each Forest on the battlefield."
//
// Every Forest counts, whoever controls it, counted once as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f3ef680-90f4-487f-b44b-db25e29e57ce",
		Name:         "Fruition",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: len(effects.MatchingBattlefield(ctx, effects.Subtype("Forest")))}.Apply(ctx)
		},
	})
}
