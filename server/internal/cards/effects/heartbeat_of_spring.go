package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartbeat of Spring — Enchantment {2}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Whenever a player taps a land for mana, that player adds one mana
//	 of any type that land produced."
//
// Mana Flare in green: the same triggered mana ability on the same
// payload (ManaProduced.Colors), so a land that made more than one type
// offers one of them, and the extra mana carries none of the land's
// restrictions or riders (the rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f15911a5-f597-452e-a015-b115339f28fc",
		Name:         "Heartbeat of Spring",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverAPlayerTapsALandForMana(
				"Heartbeat of Spring — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
	})
}
