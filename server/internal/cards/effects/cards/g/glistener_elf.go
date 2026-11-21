package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glistener Elf — Creature — Phyrexian Elf Warrior {G}, 1/1 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Infect is the engine's damage tail (CR 702.90).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d95d173-5c7f-4e0c-bcdc-9b90fcd7339b",
		Name:            "Glistener Elf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
