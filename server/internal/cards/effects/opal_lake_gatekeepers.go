package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Opal Lake Gatekeepers — Creature — Vedalken Soldier {3}{U}, 2/4:
//
//	"When this creature enters, if you control two or more Gates, you
//	 may draw a card."
//
// An intervening "if" (CR 603.4), Duergar Hedge-Mage's shape: the Gate
// count is checked as the creature enters (no trigger, and no target,
// otherwise) and again as the trigger resolves. A Gate is a land subtype,
// so a land that is a Gate by an effect counts.
//
// The "may" is a CR 603.5 choice asked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5375366-7ddc-4bd5-91c9-dc08f8beb3f6",
		Name:         "Opal Lake Gatekeepers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(InterveningIf(condControlTwoOrMoreOfSubtype("Gate"),
				On(game.EventETB, AllOf(Self, controlTwoOrMoreOfSubtype("Gate")),
					"Opal Lake Gatekeepers — if you control two or more Gates, you may draw a card",
					Do(DrawCards{N: 1}))),
				"Opal Lake Gatekeepers — draw a card?", 1),
		},
	})
}
