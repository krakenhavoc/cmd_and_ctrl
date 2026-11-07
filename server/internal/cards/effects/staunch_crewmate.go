package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staunch Crewmate — Creature — Human Pirate {1}{U}, 2/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, look at the top four cards of your
//	 library. You may reveal an artifact or Pirate card from among them
//	 and put it into your hand. Put the rest on the bottom of your library
//	 in a random order."
//
// Horn of the Mark's sentence (LookAtTopThenMayTakeToHand) with an
// artifact-or-Pirate filter: only the controller sees the four, the card
// taken is revealed, and the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	const label = "Staunch Crewmate — reveal an artifact or Pirate card and put it into your hand"
	Register(Spec{
		OracleID:     "8ab49974-3f06-4b6b-a416-d6f872b3887f",
		Name:         "Staunch Crewmate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label,
				LookAtTopThenMayTakeToHand(4, Or(Artifact(), HasSubtype("Pirate")), 1, label)),
		},
	})
}
