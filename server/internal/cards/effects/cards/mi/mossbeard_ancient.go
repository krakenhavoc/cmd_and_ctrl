package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mossbeard Ancient — Creature — Treefolk {5}{G}{G}, 7/7 (Dominaria United, S82):
//
//	"Trample
//	 When this creature enters, you gain 5 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "26afc8c1-c23f-4212-a2aa-386c71c4a7f9",
		Name:            "Mossbeard Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mossbeard Ancient — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
}
