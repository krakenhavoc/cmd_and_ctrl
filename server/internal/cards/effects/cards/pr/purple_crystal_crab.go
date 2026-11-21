package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purple-Crystal Crab — Creature — Crab {1}{U}, 1/1 (Foundations
// Jumpstart):
//
//	"When this creature dies, draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df4569d7-ee75-4ba0-bbc3-7766ac686523",
		Name:         "Purple-Crystal Crab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Purple-Crystal Crab — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
