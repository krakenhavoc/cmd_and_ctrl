package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eye of Jace — Artifact {1}:
//
//	"At the beginning of your upkeep, surveil 1. Then if there are
//	 seven or more cards in your graveyard, sacrifice this artifact, it
//	 deals 2 damage to each opponent, and you gain 2 life."
//
// An upkeep trigger whose "then" half rides Surveil.Then, so the
// graveyard is counted after the player has decided whether the top
// card goes there (rfMiscAEyeOfJaceUpkeep). The "if" is not an
// intervening one: the trigger always goes on the stack and the count
// is taken as the surveil settles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8f802ad-4621-4f71-b52d-fbc55853dc88",
		Name:         "Eye of Jace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Eye of Jace — surveil 1, then maybe sacrifice it for 2 damage and 2 life", rfMiscAEyeOfJaceUpkeep),
		},
	})
}
