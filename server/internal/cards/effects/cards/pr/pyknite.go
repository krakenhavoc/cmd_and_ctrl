package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyknite — Creature — Ouphe {2}{G}, 1/1 (Ice Age, S82 OLD batch 837, #5636):
//
//	"When this creature enters, draw a card at the beginning of the next turn's
//	 upkeep."
//
// Carrier Pigeons' clause: the enters trigger uses the stack (CR 603.6a) and
// schedules the delayed draw (CR 603.7a, next_upkeep_draw.go); the draw happens
// even if Pyknite has left by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "857f9a8d-ba77-4afd-922d-994f66e478ef",
		Name:         "Pyknite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pyknite — draw a card at the beginning of the next turn's upkeep",
				effects.Do(effects.DrawACardAtTheNextUpkeep("Pyknite — draw a card"))),
		},
	})
}
