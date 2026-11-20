package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Floodhound — Creature — Elemental Dog {U}, 1/2 (Modern Horizons 2,
// #5501):
//
//	"{3}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// The Clue taplands' investigate on a creature: the {T} waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6dd8ba09-f117-4514-9e09-f9f9f5fb5025",
		Name:         "Floodhound",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Investigate.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Effect:  Do(Investigate{N: 1}),
		}},
	})
}
