package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quicksilver Fisher — Creature — Phyrexian Drake {3}{U}{U}, 4/3:
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// Bellowing Crier's loot: the discard prompt opens after the draw, so the
// drawn card is a legal discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3e3d037-992b-4718-913d-7ab509d2047e",
		Name:            "Quicksilver Fisher",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Quicksilver Fisher — draw a card, then discard a card", b35LootOne),
		},
	})
}
