package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riptide Crab — Creature — Crab {1}{W}{U}, 1/3 (Invasion, #5616):
//
//	"Vigilance
//	 When this creature dies, draw a card."
//
// Vigilance is a printed keyword; the dies trigger is Buzz Bots' draw
// for its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ae8a583-9843-474a-bd46-8090df883621",
		Name:            "Riptide Crab",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Riptide Crab — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
