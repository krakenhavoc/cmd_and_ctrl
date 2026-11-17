package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bog Badger — Creature — Badger {2}{G}, 3/3 (Dominaria United, #5487):
//
//	"Kicker {B} (You may pay an additional {B} as you cast this spell.)
//	 When this creature enters, if it was kicked, creatures you control
//	 gain menace until end of turn."
//
// The creatures are fixed as the trigger resolves (CR 611.2c): one that
// enters later this turn does not gain menace. Along the Crooked Way's
// grant, behind the kicked intervening "if" (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "6d8e464c-9df1-46b1-88e6-dcc7fe7b7caa",
		Name:          "Bog Badger",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{B}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Bog Badger — kicked: creatures you control gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(YouControl(), Creature()),
						Keywords: []string{"menace"},
						Label:    "Bog Badger — menace",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
