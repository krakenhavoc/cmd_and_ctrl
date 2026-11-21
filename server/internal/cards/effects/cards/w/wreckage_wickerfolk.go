package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wreckage Wickerfolk — Artifact Creature — Scarecrow {1}{B}, 1/3
// (Aetherdrift, #4464):
//
//	"Flying
//	 When this creature enters, surveil 2. (Look at the top two cards of
//	 your library, then put any number of them into your graveyard and
//	 the rest on top of your library in any order.)"
//
// Imperious Inkmage's enters trigger with flying in place of vigilance:
// surveil 2 (CR 701.25a) on the stack, one prompt to its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ba5af76-2485-4fd6-bd2e-fe90d1f2691c",
		Name:            "Wreckage Wickerfolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wreckage Wickerfolk — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
