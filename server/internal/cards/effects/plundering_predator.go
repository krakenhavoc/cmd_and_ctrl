package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plundering Predator — Creature — Dragon {4}{R}, 4/4 (Jumpstart 2022,
// #5948):
//
//	"Flying
//	 When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// Azra Bladeseeker's optional loot: the discard is asked on resolution and
// only a card actually discarded draws.
//
// No simplification.
func init() {
	const label = "Plundering Predator — you may discard a card. If you do, draw a card"
	Register(Spec{
		OracleID:        "dfd260f2-90e6-464b-91e2-a66bb3877918",
		Name:            "Plundering Predator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{WhenThisEnters(label, gapB157MayDiscardThenDraw(label))},
	})
}
