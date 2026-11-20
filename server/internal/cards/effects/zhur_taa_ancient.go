package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zhur-Taa Ancient — Creature — Beast {3}{R}{G}, 7/5:
//
//	"Whenever a player taps a land for mana, that player adds one mana
//	 of any type that land produced."
//
// Mana Flare's printed sentence on a 7/5 body: a mana trigger (CR 605.1b)
// for every player's land, adding one more mana of a type that land just
// produced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6543b60c-12b0-4e69-9074-6a2b468c5234",
		Name:         "Zhur-Taa Ancient",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverAPlayerTapsALandForMana(
				"Zhur-Taa Ancient — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
	})
}
