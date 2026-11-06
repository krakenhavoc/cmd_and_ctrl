package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bark-Knuckle Boxer — Creature — Raccoon Berserker {1}{G}, 3/2
// (Bloomburrow, #4467):
//
//	"Whenever you expend 4, this creature gains indestructible until end
//	 of turn. (You expend 4 as you spend your fourth total mana to cast
//	 spells during a turn.)"
//
// Expend is WheneverYouExpend (expend.go); the grant is
// thisCreatureUntilEOT with no pump, which does nothing if the Boxer has
// left by the time the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "08e173e3-4c28-49d2-bb57-ccf533f20daf",
		Name:           "Bark-Knuckle Boxer",
		ReadsManaSpent: game.ReadsYourCasts,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouExpend(4, "Bark-Knuckle Boxer — this creature gains indestructible until end of turn",
				thisCreatureUntilEOT("Bark-Knuckle Boxer — indestructible", 0, 0, "indestructible")),
		},
	})
}
