package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Echo Inspector — Creature — Bird Rogue {3}{U}, 2/3 (Streets of New Capenna, #5491):
//
//	"Flying
//	 When this creature enters, it connives. (Draw a card, then discard a
//	 card. If you discarded a nonland card, put a +1/+1 counter on this
//	 creature.)"
//
// The enters trigger is connive.go's ConniveEach for the one creature that
// entered (WhenThisEntersItConnives): draw, then the controller's own
// discard, and the +1/+1 counter only for a nonland card (CR 701.50a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d7ae0009-3747-4f3a-a613-3e640b0219a0",
		Name:            "Echo Inspector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersItConnives("Echo Inspector — it connives"),
		},
	})
}
