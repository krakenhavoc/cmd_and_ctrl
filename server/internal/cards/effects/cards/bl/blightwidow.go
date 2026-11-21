package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blightwidow — Creature — Phyrexian Spider {3}{G}, 2/4 (Mirrodin Besieged,
// #5576):
//
//	"Reach (This creature can block creatures with flying.)
//	 Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Reach is read at the block declaration (CR 702.17b) and infect is the
// engine's damage tail (CR 702.90b): -1/-1 counters to creatures, poison to
// players.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "282003f4-862a-4f80-8f45-e470fc0fe2fe",
		Name:            "Blightwidow",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "infect"},
	})
}
