package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoke — Enchantment {R}{R}:
//
//	"Players can't untap more than one creature during their untap
//	 steps."
//
// Static Orb's and Winter Orb's cap (#826, ADR 0070) over creatures,
// with no "as long as it is untapped" condition: each player's untap
// step offers their tapped creatures and takes at most one (CR 502.3).
// Non-creature permanents untap as normal and are never offered. It
// composes with the orbs, which is the point of the one-predicate
// solver.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8aa97d25-cd51-4ceb-b7eb-af64f0914a8c",
		Name:         "Smoke",
		Completeness: effects.CompletenessFull,
		UntapCaps: []game.UntapCap{
			effects.CantUntapMoreThan("Smoke — no more than one creature", 1, effects.Creature()),
		},
	})
}
