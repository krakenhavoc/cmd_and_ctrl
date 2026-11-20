package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raise Dead — Sorcery {B} (Portal, #5630):
//
//	"Return target creature card from your graveyard to your hand."
//
// Reaping the Graves without the storm.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbc9c731-181a-4f00-a7b0-eb7e56eac2ea",
		Name:         "Raise Dead",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
