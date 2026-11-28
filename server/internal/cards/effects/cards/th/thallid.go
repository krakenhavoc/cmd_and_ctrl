package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thallid — Creature — Fungus {G}, 1/1 (Fallen Empires, S82 OLD b853, #5637):
//
//	"At the beginning of your upkeep, put a spore counter on this creature.
//	 Remove three spore counters from this creature: Create a 1/1 green
//	 Saproling creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46abaabe-7015-4d82-a5aa-2706b1f51bed",
		Name:         "Thallid",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB853SporeUpkeep("Thallid")},
		Activated:    []effects.ActivatedAbility{effects.S82OldB853SaprolingRow()},
	})
}
