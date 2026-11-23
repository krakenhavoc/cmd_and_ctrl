package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teroh's Faithful — Creature — Human Cleric {3}{W}, 1/4 (Torment,
// #5612):
//
//	"When this creature enters, you gain 4 life."
//
// Migrating Ketradon's enters trigger without the reach and the cycling:
// a real trigger on the stack, so it can be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "875c96ce-f6e4-438b-929d-de53afdd99a1",
		Name:         "Teroh's Faithful",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Teroh's Faithful — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
