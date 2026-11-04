package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Casey Jones, Jury-Rig Justiciar — {1}{R} Legendary Creature — Human
// Berserker 2/1 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Haste
//	 When Casey Jones enters, look at the top four cards of your library.
//	 You may reveal an artifact card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order."
//
// Ingenious Smith's enters trigger: only Casey's controller sees the four,
// the card taken is revealed (CR 701.20a), and the rest go to the bottom in a
// random order once the answer is in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b4c4fdb9-e034-49c0-9f7f-07e8cae19eb2",
		Name:            "Casey Jones, Jury-Rig Justiciar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Casey Jones, Jury-Rig Justiciar — look at the top four cards; you may reveal an artifact card and put it into your hand",
				LookAtTopThenMayTakeToHand(4, Artifact(), 1, "Casey Jones — reveal an artifact card and put it into your hand")),
		},
	})
}
