package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of Epityr — Creature — Human Wizard {U}, 1/1 (Time Spiral, #5596):
//
//	"When this creature enters, look at the top four cards of your
//	 library, then put them back in any order."
//
// Sage Owl's trigger without the wings: LookAtTop over four cards, put
// back in the order its controller picks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef893585-246c-4028-bd2b-14ac51ce29be",
		Name:         "Sage of Epityr",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sage of Epityr — look at the top four cards of your library, then put them back in any order", Do(LookAtTop{N: 4})),
		},
	})
}
