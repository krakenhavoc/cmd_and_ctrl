package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disentomb — Sorcery {B} (Magic 2010, #5585):
//
//	"Return target creature card from your graveyard to your hand."
//
// Raise Dead, word for word: the card returns if it is still in your
// graveyard as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4ab986d-662c-4879-b6f1-6887a2ec4d7c",
		Name:         "Disentomb",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
