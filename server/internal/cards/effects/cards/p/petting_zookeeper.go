package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Petting Zookeeper — Creature — Elf Employee {2}{G}:
//
//	"Reach
//	 When this creature enters, open an Attraction."
//
// Opening an Attraction is the engine's (ADR 0207,
// game.OpenAttractionsForEffect): the top card of its controller's
// Attraction deck enters under their control. A player playing without an
// Attraction deck, or with an empty one, opens nothing (CR 701.51a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad6dd363-e1b7-4df4-b9b6-4f5c1fae3bc6",
		Name:            "Petting Zookeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{effects.WhenThisEntersOpenAnAttraction("Petting Zookeeper")},
	})
}
