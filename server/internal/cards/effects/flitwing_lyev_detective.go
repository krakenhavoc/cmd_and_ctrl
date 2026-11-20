package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flitwing, Lyev Detective — Legendary Creature — Faerie Detective {2}{U},
// 1/1 (Mystery Booster Commander Edition):
//
//	"Flying
//	 If you would create one or more tokens, you may create that many Clue
//	 tokens instead."
//
// The creation event is a list of groups, each a template and a count
// (CR 701.7b). The replacement rewrites EVERY group into a Clue group with
// the same count, so "that many" holds per group and two kinds of token
// become that many Clues in all. It is a "may": a yes/no prompt is asked
// each time, and "no" leaves the event alone. Only the controller's own
// creations are watched.
//
// No simplification.
func init() {
	rep := hobA9InsteadCreateTokens("Flitwing, Lyev Detective: create that many Clue tokens instead",
		func(game.Card) bool { return true },
		func() []game.Card { return []game.Card{ClueToken()} })
	rep.Optional = true
	rep.PromptQuestion = "Flitwing, Lyev Detective: create that many Clue tokens instead?"
	Register(Spec{
		OracleID:     "e8a9dddd-4527-46cd-8b7e-4f83c5b51e48",
		Name:         "Flitwing, Lyev Detective",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{rep},
	})
}
