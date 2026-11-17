package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiritual Guardian — Creature — Spirit {3}{W}{W}, 3/4 (Portal, #5630):
//
//	"When this creature enters, you gain 4 life."
//
// An enters trigger on the stack, like Mandroid Squadron's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d03860b4-c663-4230-9200-6b89fa849ae7",
		Name:         "Spiritual Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spiritual Guardian — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
