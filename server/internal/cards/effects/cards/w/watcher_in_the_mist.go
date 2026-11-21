package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watcher in the Mist — Creature — Spirit {3}{U}{U}, 3/4 (Ravnica: Clue Edition):
//
//	"Flying
//	 When this creature enters, surveil 2."
//
// A.I.M. Synthoids's ETB trigger: the surveil goes on the stack, so an
// opponent can respond to it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8c0fbea-a678-46b0-b385-e36f14c07672",
		Name:            "Watcher in the Mist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Watcher in the Mist — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
	})
}
