package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Disentomb — Sorcery {B} (Magic 2010, #5585):
//
//	"Return target creature card from your graveyard to your hand."
//
// Raise Dead, word for word: the card returns if it is still in your
// graveyard as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4ab986d-662c-4879-b6f1-6887a2ec4d7c",
		Name:         "Disentomb",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
