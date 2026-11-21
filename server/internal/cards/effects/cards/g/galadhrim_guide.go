package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galadhrim Guide — Creature — Elf Scout {3}{G}, 3/4 (The Hobbit Eternal,
// HOC):
//
//	"When this creature enters, scry 2."
//
// Faerie Seer's trigger without the flying (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9204de14-cbb6-40e5-9dcc-44cce3723e8e",
		Name:         "Galadhrim Guide",
		Purpose:      game.Purpose{Scry: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Galadhrim Guide — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
