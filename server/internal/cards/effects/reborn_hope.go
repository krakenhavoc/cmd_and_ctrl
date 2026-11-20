package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reborn Hope — Sorcery {G}{W} (Alara Reborn, #5586):
//
//	"Return target multicolored card from your graveyard to your hand."
//
// Raise Dead with the multicolored test (two or more colours, Jared
// Carthalion's −6 clause) in place of the creature one. A card that left
// the graveyard in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "683d8deb-2ab4-4c65-9a23-f4e733f06ba2",
		Name:         "Reborn Hope",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("target multicolored card from your graveyard", YouOwn(), Multicolored()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
