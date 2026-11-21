package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Priests of Norn — Creature — Phyrexian Cleric {2}{W}, 1/4 (Mirrodin Besieged,
// #5576):
//
//	"Vigilance
//	 Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Vigilance and infect are printed keywords the engine reads in combat
// (CR 702.20b, 702.90b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "793fe31a-2b11-49f8-9851-719469ad1726",
		Name:            "Priests of Norn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "infect"},
	})
}
