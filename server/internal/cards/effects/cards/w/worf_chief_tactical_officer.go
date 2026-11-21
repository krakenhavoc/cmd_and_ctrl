package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Worf, Chief Tactical Officer — Legendary Creature — Klingon Warrior
// Officer {2}{W}, 2/3 (Star Trek Commander, #4974):
//
//	"Vigilance"
//
// The printed keyword, which the engine enforces (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "26e86be4-52e1-47cb-b548-c434482bd3ad",
		Name:            "Worf, Chief Tactical Officer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
