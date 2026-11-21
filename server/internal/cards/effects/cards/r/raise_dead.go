package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raise Dead — Sorcery {B} (Portal, #5630):
//
//	"Return target creature card from your graveyard to your hand."
//
// Reaping the Graves without the storm.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbc9c731-181a-4f00-a7b0-eb7e56eac2ea",
		Name:         "Raise Dead",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
