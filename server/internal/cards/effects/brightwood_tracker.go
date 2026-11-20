package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightwood Tracker — Creature — Elf Scout {3}{G}, 2/4 (Core Set 2020, #5518):
//
//	"{5}{G}, {T}: Look at the top four cards of your library. You may reveal a
//	 creature card from among them and put it into your hand. Put the rest
//	 on the bottom of your library in a random order."
//
// Squad Rallier's look-and-take (LookAtTopThenMayTakeToHand) for any
// creature card, with a {T} in the cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8d44858-1d21-43fd-b193-59ddf7c53785",
		Name:         "Brightwood Tracker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{5}{G}, {T}: Look at the top four cards of your library. You may reveal a creature card " +
				"from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    Plus(ManaCost("{5}{G}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: LookAtTopThenMayTakeToHand(4, Creature(), 1,
				"Brightwood Tracker — you may reveal a creature card and put it into your hand"),
		}},
	})
}
