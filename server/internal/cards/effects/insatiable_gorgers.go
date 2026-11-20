package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insatiable Gorgers — Creature — Vampire Berserker {2}{R}{R}, 5/3
// (Eldritch Moon, #5541):
//
//	"This creature attacks each combat if able.
//	 Madness {3}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Bloodrock Cyclops's attack requirement (CR 508.1d) and the one Madness
// string (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4534f33-8784-40a5-af1a-0f9975741aa2",
		Name:         "Insatiable Gorgers",
		Completeness: CompletenessFull,
		Madness:      "{3}{R}",
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
