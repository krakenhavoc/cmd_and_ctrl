package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Freestrider Lookout — Creature — Human Rogue {2}{G}, 3/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Reach
//	 Whenever you commit a crime, look at the top five cards of your
//	 library. You may put a land card from among them onto the
//	 battlefield tapped. Put the rest on the bottom of your library in a
//	 random order. This ability triggers only once each turn."
//
// Hardbristle Bandit's once-each-turn crime trigger (CR 700.13) with
// Kaslem's Stonetree's look (lookAtTopMayPutALandTapped) over five cards
// instead of six.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b7d3e234-0954-4422-96c0-cda91101faa3",
		Name:            "Freestrider Lookout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCommitACrimeOnceEachTurn("Freestrider Lookout — look at the top five; you may put a land onto the battlefield tapped (crime)",
				lookAtTopMayPutALandTapped(5, "Freestrider Lookout — you may put a land card from among them onto the battlefield tapped")),
		},
	})
}
