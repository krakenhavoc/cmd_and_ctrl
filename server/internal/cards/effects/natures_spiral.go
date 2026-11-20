package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "565922d6-8152-411f-8d4d-dc7c8523a49b",
		Name:         "Nature's Spiral",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
