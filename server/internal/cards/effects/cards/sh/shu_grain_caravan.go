package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Grain Caravan — Creature — Human Soldier {2}{W}, 2/2 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"When this creature enters, you gain 2 life."
//
// An enters trigger on the stack (Balamb T-Rexaur's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "587e9f15-1d79-458f-a073-229b828b2393",
		Name:         "Shu Grain Caravan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shu Grain Caravan — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
