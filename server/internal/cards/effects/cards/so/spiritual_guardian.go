package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiritual Guardian — Creature — Spirit {3}{W}{W}, 3/4 (Portal, #5630):
//
//	"When this creature enters, you gain 4 life."
//
// An enters trigger on the stack, like Mandroid Squadron's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d03860b4-c663-4230-9200-6b89fa849ae7",
		Name:         "Spiritual Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spiritual Guardian — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
