package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fatigue — Sorcery {1}{U} (Urza's Destiny, S82 #5621):
//
//	"Target player skips their next draw step."
//
// Ivory Gargoyle's one-shot skip record (Game.SkipNextStepForEffect, #6144,
// CR 614.10a) on the target player: spent by the next draw step that player
// would take, even one that is several turns away.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a88dcb6-a391-408f-8bfc-7b5b2cc34267",
		Name:         "Fatigue",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := effects.TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			ctx.Game.SkipNextStepForEffect(player, game.StepDraw, item.SourceCardID, "Fatigue — that player skips their next draw step")
			return nil
		},
	})
}
