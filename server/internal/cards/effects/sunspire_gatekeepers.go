package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunspire Gatekeepers — Creature — Human Soldier {3}{W}, 2/4:
//
//	"When this creature enters, if you control two or more Gates,
//	 create a 2/2 white Knight creature token with vigilance."
//
// An intervening "if" (CR 603.4), Duergar Hedge-Mage's shape: the Gate
// count is checked as the creature enters (no trigger, and no target,
// otherwise) and again as the trigger resolves. A Gate is a land subtype,
// so a land that is a Gate by an effect counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df2cc4ba-70b8-4875-8f84-9c268ab8d5f6",
		Name:         "Sunspire Gatekeepers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(condControlTwoOrMoreOfSubtype("Gate"),
				On(game.EventETB, AllOf(Self, controlTwoOrMoreOfSubtype("Gate")),
					"Sunspire Gatekeepers — if you control two or more Gates, create a 2/2 white Knight creature token with vigilance",
					Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1}))),
		},
	})
}
