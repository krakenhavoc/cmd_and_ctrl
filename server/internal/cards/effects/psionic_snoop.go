package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Psionic Snoop — Creature — Human Rogue {2}{U}, 0/3 (Streets of New Capenna, #5491):
//
//	"Flash
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
		OracleID:        "a4e8d2ad-cce2-45de-b9a2-7ba3b1546efb",
		Name:            "Psionic Snoop",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersItConnives("Psionic Snoop — it connives"),
		},
	})
}
