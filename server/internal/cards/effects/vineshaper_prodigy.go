package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vineshaper Prodigy — Creature — Elf Druid {1}{G}, 2/2 (Dominaria
// United, #5487):
//
//	"Kicker {1}{U} (You may pay an additional {1}{U} as you cast this
//	 spell.)
//	 When this creature enters, if it was kicked, look at the top three
//	 cards of your library. Put one of them into your hand and the rest on
//	 the bottom of your library in any order."
//
// Raven Familiar's look-three-take-one-rest-on-the-bottom behind the
// kicked intervening "if" (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c41172c6-5133-48ed-bde0-3168b5371c18",
		Name:          "Vineshaper Prodigy",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Vineshaper Prodigy — kicked: look at the top three, take one",
				LookAtTopThenTakeOneRestOnBottom(3, "Vineshaper Prodigy — put one into your hand")),
		},
	})
}
