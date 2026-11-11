package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7e6c7b57-bd6a-4ac8-bcab-c3a879f479c2",
		Name:         "Chatterstorm",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 1}.Apply(ctx)
		},
	})
}
