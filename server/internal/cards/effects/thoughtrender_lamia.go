package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thoughtrender Lamia — Enchantment Creature — Lamia {4}{B}{B}, 5/3 (Journey into Nyx, #5558):
//
//	"Constellation — Whenever this creature or another enchantment you
//	 control enters, each opponent discards a card."
//
// Constellation is Eidolon of Blossoms' condition: its own entry (it is an
// enchantment) or another enchantment entering under your control. Each
// opponent discards a card of their choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d97c4bca-6433-4cb0-b63a-fdf8360fc29c",
		Name:         "Thoughtrender Lamia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, constellationTriggers, "Thoughtrender Lamia — each opponent discards a card (constellation)", eachOpponentDiscardsOne),
		},
	})
}
