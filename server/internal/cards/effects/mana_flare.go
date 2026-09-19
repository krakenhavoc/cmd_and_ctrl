package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mana Flare — Enchantment {2}{R} (#394):
//
//	"Whenever a player taps a land for mana, that player adds one mana
//	 of any type that land produced."
//
// The card that needed the PRODUCED COLOUR, which is the second half
// of #763 and the engine-seams row this closes alongside it: before
// ManaProduced.Colors nothing could tell a trigger which type the land
// had just made, so "one mana of any type that land produced" was
// unwritable however the trigger itself was modelled.
//
// Three details the printed card actually turns on, all of them
// carried by the payload rather than by this file:
//
//   - "A PLAYER", not you. It fires for every player's land, which is
//     what makes Mana Flare a group-hug card and a mistake to play.
//     The mana goes to whoever tapped the land.
//   - "ANY TYPE THAT LAND PRODUCED", not any colour. A Forest offers
//     {G} and nothing else, so no prompt is owed at all; a dual that
//     produced one type offers that one type. Colourless counts as a
//     type, so a Wastes tapped under a Mana Flare makes a second {C}.
//   - The mana is ADDITIONAL and lands with no priority window (CR
//     605.4a), so it is spendable on the spell the land was tapped for.
//
// The land's own colour pick is where the type becomes known for a
// dual, and the trigger fires from there (ADR 0074 §3) — so a Mana
// Flare over a Thriving Isle sees exactly the colour the player chose.
//
// Same declared auto-tap simplification as Wild Growth: the planner
// does not count the extra mana, so an auto-tapped cast under a Mana
// Flare taps as many lands as it would without one and floats the
// difference (ADR 0074 §7).
//
// No simplification of the card itself.
func init() {
	Register(Spec{
		OracleID:     "97159138-c34b-416e-b079-5c952383a243",
		Name:         "Mana Flare",
		Completeness: CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			WheneverAPlayerTapsALandForMana(
				"Mana Flare — add one mana of any type that land produced",
				AddsOneManaOfAnyTypeProduced()),
		},
	})
}
