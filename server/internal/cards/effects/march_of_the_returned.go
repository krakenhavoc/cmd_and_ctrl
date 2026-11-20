package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March of the Returned — Sorcery {3}{B} (Theros, #5561):
//
//	"Return up to two target creature cards from your graveyard to your hand."
//
// Fight On!'s body: up to two creature cards from your own graveyard,
// each returned while it is still a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53a40369-0e62-4ad2-a6bd-65a5ff0da65e",
		Name:         "March of the Returned",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve:    ResolveSteps(Instruction(returnLegalGraveyardTargetsToHand)),
	})
}
