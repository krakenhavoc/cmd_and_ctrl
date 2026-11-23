package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "6d8e464c-9df1-46b1-88e6-dcc7fe7b7caa",
		Name:          "Bog Badger",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Bog Badger — kicked: creatures you control gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.YouControl(), effects.Creature()),
						Keywords: []string{"menace"},
						Label:    "Bog Badger — menace",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
