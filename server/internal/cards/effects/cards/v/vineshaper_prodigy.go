package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "c41172c6-5133-48ed-bde0-3168b5371c18",
		Name:          "Vineshaper Prodigy",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Vineshaper Prodigy — kicked: look at the top three, take one",
				effects.LookAtTopThenTakeOneRestOnBottom(3, "Vineshaper Prodigy — put one into your hand")),
		},
	})
}
