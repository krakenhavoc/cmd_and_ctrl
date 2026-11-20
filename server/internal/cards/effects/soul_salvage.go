package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul Salvage — Sorcery {2}{B} (Dominaria, #5530):
//
//	"Return up to two target creature cards from your graveyard to your hand."
//
// Fight On!'s sentence: each target creature card still in your
// graveyard as it resolves returns to your hand (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60b6bd93-3eb3-4b81-bc4d-1e3351ee2e64",
		Name:         "Soul Salvage",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve:    ResolveSteps(Instruction(returnLegalGraveyardTargetsToHand)),
	})
}
