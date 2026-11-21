package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chatterstorm — Sorcery {1}{G}:
//
//	"Create a 1/1 green Squirrel creature token.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn.)"
//
// Empty the Warrens' shape with one Squirrel: each storm copy is its
// own resolution that creates its own token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e6c7b57-bd6a-4ac8-bcab-c3a879f479c2",
		Name:         "Chatterstorm",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Squirrel"), N: 1}.Apply(ctx)
		},
	})
}
