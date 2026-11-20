package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screeching Drake — Creature — Drake {3}{U}, 2/2 (Portal Second Age,
// #5625):
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// Icewind Elemental's text: the entry loot (b35LootOne) on a flier, and
// flying is printed data (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "da2ac9fa-52c2-4896-a596-89f287406641",
		Name:            "Screeching Drake",
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Screeching Drake — draw a card, then discard a card", b35LootOne),
		},
	})
}
