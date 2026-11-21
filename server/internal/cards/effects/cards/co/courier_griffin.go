package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courier Griffin — Creature — Griffin {3}{W}, 2/3 (Battle for
// Zendikar, #5547):
//
//	"Flying
//	 When this creature enters, you gain 2 life."
//
// Kemba's Skyguard's text: flying carried as printed data (CR 702.9b),
// and the entry trigger uses the stack, so it can be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2bb0a5e3-ccb0-4f87-b021-e60ebfa35b2b",
		Name:            "Courier Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{LifeGain: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Courier Griffin — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
