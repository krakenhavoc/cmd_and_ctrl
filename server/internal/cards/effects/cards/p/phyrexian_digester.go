package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Digester — Artifact Creature — Phyrexian Construct {3}, 2/1 (Mirrodin Besieged,
// #5576):
//
//	"Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Infect is the engine's damage tail (CR 702.90b): -1/-1 counters to
// creatures, poison counters to players.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9e7b03ae-dcf5-4892-9000-aee7e2cd2fed",
		Name:            "Phyrexian Digester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
