package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nesting Wurm — Creature — Wurm {4}{G}{G}, 4/3 (Nemesis, S82 OLD, #5618):
//
//	"Trample
//	 When this creature enters, you may search your library for up to three
//	 cards named Nesting Wurm, reveal them, put them into your hand, then
//	 shuffle."
//
// The search names a card name (CR 701.23b), so it may find fewer than three
// or none, and the "you may" lets the controller decline it and the shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3160dada-d59f-4f23-8420-ba14f9df7514",
		Name:            "Nesting Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nesting Wurm — search for up to three cards named Nesting Wurm",
				effects.S82OldB677SearchForUpToThreeNamed("Nesting Wurm")),
		},
	})
}
