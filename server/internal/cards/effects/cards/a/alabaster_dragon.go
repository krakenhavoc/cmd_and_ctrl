package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alabaster Dragon — Creature — Dragon {4}{W}{W}, 4/4 (Weatherlight, S82 older sets batch 777, #5629):
//
//	"Flying
//	 When this creature dies, shuffle it into its owner's library."
//
// Flying rides PrintedKeywords. Angel of Fury's shuffle without the "may": the card goes into
// its owner's library if it is still in the graveyard as the card that died (CR 400.7), and that
// library is shuffled whether or not it did (CR 701.24c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2392a41a-59d3-4749-be94-4d9df0af9c4c",
		Name:            "Alabaster Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Alabaster Dragon — shuffle it into its owner's library", effects.S82OldB734ShuffleDiedCardIntoLibrary),
		},
	})
}
