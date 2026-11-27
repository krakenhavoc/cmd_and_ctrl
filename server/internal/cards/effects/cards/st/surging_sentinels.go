package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surging Sentinels — Creature — Human Soldier {2}{W}, 2/1 (Coldsnap, #5597):
//
//	"First strike
//	 Ripple 4 (...)"
//
// First strike with ripple (CR 702.60a) as a trigger of the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b98ae2a-56ac-485d-b064-a06e1e8d4055",
		Name:            "Surging Sentinels",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{effects.S82OldB566Ripple("Surging Sentinels", 4)},
	})
}
