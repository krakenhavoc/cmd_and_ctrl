package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dutiful Return — Sorcery {3}{B} (Khans of Tarkir, #5554):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand."
//
// Blood Fountain's return as a sorcery: up to two creature cards in your
// graveyard, each still-legal one to your hand (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8f958da-3a31-4064-9a78-66e9b65b85ab",
		Name:         "Dutiful Return",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
