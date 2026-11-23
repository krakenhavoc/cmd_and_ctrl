package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shriek Raptor — Creature — Phyrexian Bird {3}{W}{W}, 2/3 (New Phyrexia, #5575):
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Plague Stinger's two keywords; infect is the engine's damage tail
// (CR 702.90).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "426bfd00-e727-49f4-be2d-df5289d2cc61",
		Name:            "Shriek Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
	})
}
