package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mr. Spock, Science Officer — Legendary Creature — Vulcan Human
// Scientist {2}{U}{U}, 2/4 (Star Trek, #4973):
//
//	"If you would draw a card, draw two cards instead."
//
// Thought Reflection's replacement on a creature: the same
// YouDrawTwiceInstead, so each draw of a "draw N" is doubled on its own
// (CR 121.2) and two of them draw four.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "869bde03-8813-43a9-9df5-a21c62511b4b",
		Name:         "Mr. Spock, Science Officer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			YouDrawTwiceInstead("Mr. Spock, Science Officer: draw two cards instead"),
		},
	})
}
