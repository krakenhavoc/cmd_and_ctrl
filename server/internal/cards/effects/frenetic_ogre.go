package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frenetic Ogre — Creature — Ogre {4}{R}, 2/3:
//
//	"{R}, Discard a card at random: This creature gets +3/+0 until end of
//	 turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6eb4b649-c578-4b29-8da1-7c757412bec0",
		Name:         "Frenetic Ogre",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Discard a card at random: This creature gets +3/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{R}"), DiscardAtRandom(1, "a card at random")),
			Effect:  thisGetsUntilEndOfTurn(3, 0, "Frenetic Ogre — +3/+0"),
		}},
	})
}
