package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Fisher — Creature — Bird Soldier {3}{U}, 2/2 (Dominaria Remastered,
// #5948):
//
//	"Flying
//	 When this creature dies, you may draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69a72414-d02f-4a5c-b5c9-8941bb181f61",
		Name:            "Aven Fisher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.WhenThisDies("Aven Fisher — draw a card", effects.Do(effects.DrawCards{N: 1})), "Aven Fisher — draw a card?", 1),
		},
	})
}
