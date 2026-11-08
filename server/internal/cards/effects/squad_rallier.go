package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squad Rallier — Creature — Human Scout {3}{W}, 3/4 (Foundations,
// #4465):
//
//	"{2}{W}: Look at the top four cards of your library. You may reveal a
//	 creature card with power 2 or less from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// Water Tribe Rallier's look-and-take (LookAtTopThenMayTakeToHand) for a
// plain mana cost: only you see the four, the card taken is revealed, and
// the rest go to the bottom in a random order. A card's power in the
// library is its printed power. There is no {T} in the cost, so it can be
// activated any number of times, summoning sick or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0b15374-b4f6-48e0-8477-4eb5d5c2a938",
		Name:         "Squad Rallier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{2}{W}: Look at the top four cards of your library. You may reveal a creature card with power 2 or less " +
				"from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: LookAtTopThenMayTakeToHand(4, And(Creature(), PowerLE(2)), 1,
				"Squad Rallier — you may reveal a creature card with power 2 or less and put it into your hand"),
		}},
	})
}
