package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fight On! — Instant {2}{B} (Final Fantasy, #4462):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand."
//
// Pull from the Grave's returns without the life: up to two, so the spell
// can be cast with no target, and each card still in the graveyard as a
// creature card comes back; one that left in response stays where it went
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ab42daa-8726-4973-947d-1d6c32f0ba61",
		Name:         "Fight On!",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve:    ResolveSteps(Instruction(returnLegalGraveyardTargetsToHand)),
	})
}
