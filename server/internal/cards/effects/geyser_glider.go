package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geyser Glider — Creature — Elemental Beast {3}{R}{R}, 4/4 (Zendikar, #5583):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 flying until end of turn."
//
// Landfall grants flying to the creature itself until end of turn, a
// layer-6 effect pinned to the source.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "840d1002-abf4-4d05-945a-05c3fe9a2ebd",
		Name:         "Geyser Glider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Geyser Glider — gains flying until end of turn (landfall)",
				thisGainsKeywordUntilEndOfTurn("flying", "Geyser Glider — gains flying until end of turn")),
		},
	})
}
