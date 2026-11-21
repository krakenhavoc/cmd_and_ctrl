package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nature's Spiral — Sorcery {1}{G} (Magic 2010, #5585):
//
//	"Return target permanent card from your graveyard to your hand."
//
// Raise Dead's body over any permanent card: an artifact, creature,
// enchantment, land or planeswalker card in your graveyard returns to
// your hand if it is still there as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "565922d6-8152-411f-8d4d-dc7c8523a49b",
		Name:         "Nature's Spiral",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
