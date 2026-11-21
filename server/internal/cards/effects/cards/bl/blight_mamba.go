package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight Mamba — Creature — Phyrexian Snake {1}{G}, 1/1:
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)
//	 {1}{G}: Regenerate this creature."
//
// Infect is the engine's damage tail (ADR 0056); the regeneration row is
// the shared regenerate-this body (regenerate_this.go, CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e11a856c-31d3-4f4b-a517-8a5134b88517",
		Name:            "Blight Mamba",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{G}")},
	})
}
