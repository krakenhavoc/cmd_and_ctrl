package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Train of Thought — Sorcery {1}{U} (Guildpact, #5599):
//
//	"Replicate {1}{U} (When you cast this spell, copy it for each time
//	 you paid its replicate cost.)
//	 Draw a card."
//
// Replicate is Shattering Spree's pair (replicateMana and Replicate(),
// CR 702.56a), twenty payments a cast at most. The spell has no target,
// so the copies are made and resolve without a question; each draws a
// card for the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "90ba0366-57ff-4dc4-a290-3925548129ce",
		Name:          "Train of Thought",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Draws: 1},
		OptionalCosts: []game.AdditionalCost{replicateMana("{1}{U}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve:     ResolveSteps(DrawCards{N: 1}),
	})
}
