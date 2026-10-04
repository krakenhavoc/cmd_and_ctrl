package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zendikar Resurgent — Enchantment {5}{G}{G}:
//
//	"Whenever you tap a land for mana, add one mana of any type that
//	 land produced. (The types of mana are white, blue, black, red,
//	 green, and colorless.)
//	 Whenever you cast a creature spell, draw a card."
//
// The first ability is a TRIGGERED MANA ability (CR 605.1b), Mirari's
// Wake's shape: it fires on the tap, adds mana and never uses the stack
// (CR 605.4a). "Of any type that land produced" includes colorless, so
// a Wastes taps for {C}{C}: AddsOneManaOfAnyTypeProduced reads the
// land's actual production, and a land that made two types is an
// ordinary pick between them.
//
// The second is an ordinary cast trigger and goes on the stack above
// the creature spell, so it resolves first (CR 603.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0af158b-f42a-4022-a438-e0a2e1fd6d9e",
		Name:         "Zendikar Resurgent",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverYouTapALandForMana(
				"Zendikar Resurgent — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Zendikar Resurgent — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
