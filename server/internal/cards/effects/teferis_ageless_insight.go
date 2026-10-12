package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teferi's Ageless Insight — Legendary Enchantment {2}{U}{U}:
//
// "If you would draw a card except the first one you draw in each of
// your draw steps, draw two cards instead."
//
// The shared draw replacement uses the completed-draw tally for each
// step and preserves the exception when a draw is replaced by dredge.
func init() {
	Register(Spec{
		OracleID:     "1c9e1f75-73f0-4846-b53f-458a0984b1bb",
		Name:         "Teferi's Ageless Insight",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			YouDrawTwiceInsteadExceptTheFirst("Teferi's Ageless Insight: draw two cards instead"),
		},
	})
}
