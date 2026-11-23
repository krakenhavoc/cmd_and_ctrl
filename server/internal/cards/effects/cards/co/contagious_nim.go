package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contagious Nim — Creature — Phyrexian Zombie {2}{B}, 2/2 (Scars of Mirrodin, #5577)
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)"
//
// A keyword body: infect is the engine's damage tail (CR 702.90).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "02d4fe7d-ade4-40b1-b323-73a52c440525",
		Name:            "Contagious Nim",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
