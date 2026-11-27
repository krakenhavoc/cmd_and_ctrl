package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timid Drake — Creature — Drake {2}{U}, 3/3 (Mercadian Masques, #5619):
//
//	"Flying
//	 When another creature enters, return this creature to its owner's
//	 hand."
//
// Rampaging Ferocidon's "another creature" enters trigger over every player's
// creatures. The return is Coastal Wizard's BounceToHand on the Drake itself,
// which only acts on the same object that triggered (CR 400.7): a Drake that
// already left, or came back as a new object, is not returned again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa8c5467-190a-4c6a-872c-56b8f8d01a2b",
		Name:            "Timid Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEvent, "Timid Drake — return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
