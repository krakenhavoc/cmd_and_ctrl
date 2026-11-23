package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dromoka's Gift — Instant {4}{G} (Dragons of Tarkir,
// #5551):
//
//	"Bolster 4. (Choose a creature with the least toughness among creatures you control and put four +1/+1 counters on it.)"
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness is chosen as the spell resolves, and does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1264da2f-4767-4bec-b4ed-1c625ac3cf95",
		Name:         "Dromoka's Gift",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.Bolster{N: 4}),
	})
}
