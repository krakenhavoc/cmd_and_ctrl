package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frontier Seeker — Creature — Human Scout {1}{W}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"When this creature enters, look at the top five cards of your library.
//	 You may reveal a Mount creature card or a Plains card from among them
//	 and put it into your hand. Put the rest on the bottom of your library
//	 in a random order."
//
// Horn of the Mark's whole sentence (LookAtTopThenMayTakeToHand): the
// five are looked at, one matching card may be revealed and taken, and
// the rest go to the bottom at random.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "466cf331-2d9c-4ebc-84f6-30c890e1b637",
		Name:         "Frontier Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Frontier Seeker — look at the top five cards of your library",
				LookAtTopThenMayTakeToHand(5,
					Or(And(Creature(), HasSubtype("Mount")), HasSubtype("Plains")), 1,
					"Frontier Seeker — you may reveal a Mount creature card or a Plains card from among them and put it into your hand")),
		},
	})
}
