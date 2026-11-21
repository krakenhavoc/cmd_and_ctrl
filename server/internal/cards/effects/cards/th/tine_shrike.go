package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tine Shrike — Creature — Phyrexian Bird {3}{W}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Flying is read at the block declaration (CR 702.9b) and infect is the
// engine's damage tail (CR 702.90b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e20ad2f7-cb41-4e6e-8f3b-bd260f4a2ecf",
		Name:            "Tine Shrike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
	})
}
