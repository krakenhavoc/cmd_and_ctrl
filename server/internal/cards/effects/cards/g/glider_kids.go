package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glider Kids — Creature — Human Pilot Ally {2}{W}, 2/3:
//
//	"Flying
//	 When this creature enters, scry 1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ff5ce28b-25be-4136-a7b2-8f25c08a6455",
		Name:            "Glider Kids",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Glider Kids — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
