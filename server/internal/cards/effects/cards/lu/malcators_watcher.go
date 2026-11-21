package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malcator's Watcher — Artifact Creature — Phyrexian Drone {1}{U}, 1/1:
//
//	"Flying, vigilance
//	 When this creature dies, draw a card."
//
// Buzz Bots' shape: one dies trigger, which uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0493924-dc6d-4655-865f-7e9b40f547ea",
		Name:            "Malcator's Watcher",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1},
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Malcator's Watcher — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
