package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disruptor Wanderglyph — Artifact Creature — Golem {4}, 3/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Whenever this creature attacks, exile target card from an opponent's
//	 graveyard."
//
// The target is any card in a graveyard its controller does not own,
// chosen as the trigger goes on the stack (CR 603.3d); with no such card
// the trigger is removed. A card that left the graveyard in response is
// no longer a legal target and nothing is exiled (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "81dda1a8-72ac-4e69-8367-4f72c445197c",
		Name:         "Disruptor Wanderglyph",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Disruptor Wanderglyph — exile target card from an opponent's graveyard",
				exileFirstLegalCardTarget),
				TargetCardInGraveyard("target card from an opponent's graveyard", Not(YouOwn()))),
		},
	})
}
