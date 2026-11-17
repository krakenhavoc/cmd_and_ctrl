package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimson Fleet Commodore — Creature — Ogre Pirate {3}{R}, 5/2 (Commander
// Legends, #5507):
//
//	"Trample
//	 When this creature enters, you become the monarch."
//
// Trample is the printed keyword; the enters trigger is every monarch
// card's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d5ac319a-c870-4890-87c0-0e02f122f234",
		Name:            "Crimson Fleet Commodore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Crimson Fleet Commodore"),
		},
	})
}
