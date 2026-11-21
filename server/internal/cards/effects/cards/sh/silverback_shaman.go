package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverback Shaman — Creature — Ape Shaman {3}{G}{G}, 5/4 (Core Set 2020, #5518):
//
//	"Trample
//	 When this creature dies, draw a card."
//
// Buzz Bots' dies trigger with a printed trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc011947-b496-400e-99b8-b368068ba79b",
		Name:            "Silverback Shaman",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Silverback Shaman — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
