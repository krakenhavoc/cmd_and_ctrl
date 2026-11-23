package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raffine's Informant — Creature — Human Wizard {1}{W}, 2/1 (Streets of
// New Capenna, #5491):
//
//	"When this creature enters, it connives. (Draw a card, then discard a
//	 card. If you discarded a nonland card, put a +1/+1 counter on this
//	 creature.)"
//
// The enters trigger is connive.go's ConniveEach for the one creature that
// entered (WhenThisEntersItConnives): draw, then the controller's own
// discard, and the +1/+1 counter only for a nonland card (CR 701.50a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "838c4831-7d5c-4025-92b3-46eee59d6b34",
		Name:         "Raffine's Informant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersItConnives("Raffine's Informant — it connives"),
		},
	})
}
