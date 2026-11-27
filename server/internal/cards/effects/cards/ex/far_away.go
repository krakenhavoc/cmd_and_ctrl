package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Far // Away — split card, oracle e4b1ef6a-6f05-472e-aaed-1c24dd6605c7
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Far  — Instant {1}{U}: "Return target creature to its owner's hand."
//	Away — Instant {2}{B}: "Target player sacrifices a creature of their choice."
//	       Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Far under the bare oracle ID, Away under
// "#1"); the fused cast is the engine's (CR 702.102), Far resolving first. Far is
// Aether Mutation's bounce; Away is Chainer's Edict's targeted edict: the player
// chooses their own creature, hexproof or not (CR 701.21a).
//
// No simplification.
func init() {
	const away = "Away — sacrifice a creature"
	effects.Register(effects.Spec{
		OracleID:     "e4b1ef6a-6f05-472e-aaed-1c24dd6605c7",
		Name:         "Far",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BounceTheTarget,
	})
	effects.Register(effects.Spec{
		OracleID:     "e4b1ef6a-6f05-472e-aaed-1c24dd6605c7#1",
		Name:         "Away",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerSacrificesACreature(away)(ctx.Game, item)
		},
	})
}
