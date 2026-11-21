package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voda Sea Scavenger — Creature — Merfolk Rogue {2}{U}, 3/2:
//
//	"Domain — When this creature enters, look at the top X cards of your
//	 library, where X is the number of basic land types among lands you
//	 control. You may put one of those cards on top of your library. Put
//	 the rest on the bottom of your library in a random order."
//
// X is the shared domain count (DomainCount, #5666) read as the trigger
// resolves. The look is private to the controller; the card they keep stays
// on top and the rest go under in an order the game's RNG draws
// (s82DmuB10LookKeepOneOnTop). With no basic land type there is nothing to
// look at.
//
// No simplification.
func init() {
	const label = "Voda Sea Scavenger — Domain: look at the top X cards of your library"
	effects.Register(effects.Spec{
		OracleID:     "81a9da9d-c422-4293-b210-d89222633b4c",
		Name:         "Voda Sea Scavenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.S82DmuB10LookKeepOneOnTop(effects.NewContext(g, item), "Voda Sea Scavenger")
			}),
		},
	})
}
