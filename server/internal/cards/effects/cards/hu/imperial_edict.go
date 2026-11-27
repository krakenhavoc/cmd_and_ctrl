package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imperial Edict — Sorcery {1}{B}
// (Portal Three Kingdoms, #5622):
//
//	"Target opponent chooses a creature they control. Destroy that creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac2b3adb-45bb-4f56-9d52-e08771164b35",
		Name:         "Imperial Edict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.FirstLegalTargetPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.OldB711ChooseACreatureThen(ctx, player, "Imperial Edict — choose a creature you control to destroy", effects.OldB711DestroyChosen, nil)
		},
	})
}
