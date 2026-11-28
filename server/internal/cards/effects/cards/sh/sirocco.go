package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sirocco — Instant {1}{R} (Mirage, S82 OLD batch 809, #5633):
//
//	"Target player reveals their hand. For each blue instant card revealed
//	 this way, that player discards that card unless they pay 4 life."
//
// The hand is revealed, then each blue instant card in it gets its own
// pay-or-discard question (B23PayLifeUnless, CR 118.12): a player who can't
// pay (CR 119.4) or declines discards that card, if it is still in their
// hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9bdf857-0c6b-4afa-afbb-aff7e4b61b09",
		Name:         "Sirocco",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB809Sirocco,
	})
}
