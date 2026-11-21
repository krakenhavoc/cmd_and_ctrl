package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Revel Ruiner — Creature — Octopus Rogue {3}{B}, 3/1 (Streets of New Capenna, #5491):
//
//	"Menace
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
	effects.Register(effects.Spec{
		OracleID:        "adfb28ca-2ba6-4104-b731-f8fd9367f7ec",
		Name:            "Revel Ruiner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersItConnives("Revel Ruiner — it connives"),
		},
	})
}
