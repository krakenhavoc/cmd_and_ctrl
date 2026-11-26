package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alive // Well — split card, oracle 57ad3c3a-6ac7-4a35-bc07-00f6ea0988c5
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Alive — Sorcery {3}{G}: "Create a 3/3 green Centaur creature token."
//	Well  — Sorcery {W}: "You gain 2 life for each creature you control."
//	        Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Alive under the bare oracle ID, Well under
// "#1"); the fused cast is the engine's (CR 702.102), Alive resolving first, so a
// fused Well counts the Centaur. Well counts your creatures as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57ad3c3a-6ac7-4a35-bc07-00f6ea0988c5",
		Name:         "Alive",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("3/3 green Centaur"), N: 1}),
	})
	effects.Register(effects.Spec{
		OracleID:     "57ad3c3a-6ac7-4a35-bc07-00f6ea0988c5#1",
		Name:         "Well",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())))
			return effects.GainLife{Amount: 2 * n}.Apply(ctx)
		},
	})
}
