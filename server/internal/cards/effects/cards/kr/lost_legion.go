package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost Legion — Creature — Spirit Knight {1}{B}{B}, 2/3 (Throne of Eldraine, #5516):
//
//	"When this creature enters, scry 2."
//
// Automatic Librarian's shape: an enters trigger that scries two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8f3b5b5-8835-400f-a626-94ff0a007e56",
		Name:         "Lost Legion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lost Legion — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
