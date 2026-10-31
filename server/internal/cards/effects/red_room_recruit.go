package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Room Recruit — Creature — Human Spy Villain {1}{B}, 1/2 (Marvel
// Super Heroes, #4455):
//
//	"When this creature enters, it connives. (Draw a card, then discard a
//	 card. If you discarded a nonland card, put a +1/+1 counter on this
//	 creature.)"
//
// A.I.M. Scientists's enter trigger: connive.go's ConniveEach for the one
// creature that entered (CR 701.50a), the counter skipped if it has left
// (CR 701.50c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e60c38c-cbf5-4ff7-8fff-634f7e3086cc",
		Name:         "Red Room Recruit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersItConnives("Red Room Recruit — it connives"),
		},
	})
}
