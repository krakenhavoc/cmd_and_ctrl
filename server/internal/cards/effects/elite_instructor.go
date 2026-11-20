package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elite Instructor — Creature — Human Wizard {2}{U}, 2/2 (Theros Beyond Death, #5514):
//
//	"When this creature enters, draw a card, then discard a card."
//
// Temur Tawnyback's enters loot: draw a card, then discard a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6536f72e-ff4e-4cda-bc7d-909145adf829",
		Name:         "Elite Instructor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elite Instructor — draw a card, then discard a card", b35LootOne),
		},
	})
}
