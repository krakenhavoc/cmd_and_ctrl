package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Sculpt — Sorcery {1}{U} (Magic 2013, #5568):
//
//	"Target opponent mills seven cards."
//
// Dampen Thought's mill at seven, aimed at an opponent only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d8592d7-2e36-4a7a-be9e-7c8f512d5f62",
		Name:         "Mind Sculpt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerMills(7)(ctx.Game, item)
		},
	})
}
