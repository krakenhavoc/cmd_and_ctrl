package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deny Reality — Sorcery {3}{U}{B} (Alara Reborn, #5586):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast
//	 it without paying its mana cost. Put the exiled cards on the bottom in
//	 a random order.)
//	 Return target permanent to its owner's hand."
//
// Cascade is the catalog keyword (CR 702.85a). The bounce is Boomerang's
// body; a target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9e31993-bfe1-46fb-8a32-418e97f9cfe1",
		Name:         "Deny Reality",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered:    []game.TriggeredAbility{Cascade()},
		OnResolve:    bounceTheTarget,
	})
}
