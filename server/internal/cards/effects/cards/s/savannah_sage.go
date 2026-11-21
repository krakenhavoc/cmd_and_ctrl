package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savannah Sage — Creature — Cat Cleric {1}{W}, 2/2 (Core Set 2020, #5518):
//
//	"When this creature enters, you gain 2 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e6cfec7-3b54-43db-853a-27c164776a5f",
		Name:         "Savannah Sage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Savannah Sage — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
