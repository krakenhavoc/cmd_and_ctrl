package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return to Battle — Sorcery {B} (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Return target creature card from your graveyard to your hand."
//
// Raise Dead's text word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf6676b7-b40a-46bd-b166-bd3cbb49d80d",
		Name:         "Return to Battle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
