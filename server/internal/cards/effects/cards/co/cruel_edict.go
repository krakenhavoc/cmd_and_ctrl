package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cruel Edict — Sorcery {1}{B} (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Target opponent sacrifices a creature of their choice."
//
// The opponent is the target and chooses the creature they sacrifice (CR 701.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10c585c4-bf5b-4d8f-94a9-e9a5036a688f",
		Name:         "Cruel Edict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := effects.FirstLegalTarget(ctx); ok && t.Kind == game.TargetPlayer {
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					effects.SacrificeSpec("a creature", effects.Creature()),
					"Cruel Edict — sacrifice a creature")
			}
			return nil
		},
	})
}
