package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recruitment Officer — Creature — Human Soldier {W}, 2/1:
//
//	"{3}{W}: Look at the top four cards of your library. You may reveal a
//	 creature card with mana value 3 or less from among them and put it
//	 into your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// Domri's dig (LookAtTopThenMayTakeToHand): a look, so only the controller
// sees the four; the creature taken is revealed; the rest go to the bottom
// at random. The mana-value cap reads ManaValueLE.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97dfc025-afc2-42d8-a20a-f29d60754a2e",
		Name:         "Recruitment Officer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Look at the top four cards of your library. You may reveal a creature card with mana value 3 or less from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
			Cost:    effects.ManaCost("{3}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: effects.LookAtTopThenMayTakeToHand(4, effects.And(effects.Creature(), effects.ManaValueLE(3)), 1,
				"Recruitment Officer — you may reveal a creature card with mana value 3 or less and put it into your hand"),
		}},
	})
}
