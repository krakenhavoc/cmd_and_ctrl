package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Soldier-Farmers — Creature — Human Soldier {4}{W}, 2/4 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"When this creature enters, you gain 4 life."
//
// An enters trigger on the stack (Balamb T-Rexaur's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "174cf7cd-3e8c-4f90-abc1-a78a0ce832d2",
		Name:         "Shu Soldier-Farmers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shu Soldier-Farmers — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
