package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathedral Sanctifier — {W} Creature — Human Cleric 1/1 (Avacyn Restored, #5570):
//
//	"When this creature enters, you gain 3 life."
//
// Temple Acolyte's enters trigger, for 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1050f029-60d4-4769-b39f-9100c949a822",
		Name:         "Cathedral Sanctifier",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cathedral Sanctifier — gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
