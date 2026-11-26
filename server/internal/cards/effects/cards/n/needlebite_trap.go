package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needlebite Trap — Instant — Trap {5}{B}{B} (Zendikar, S82 OLD b495):
//
//	"If an opponent gained life this turn, you may pay {B} rather than pay this
//	 spell's mana cost.
//	 Target player loses 5 life and you gain 5 life."
//
// Whiplash Trap's alternative cost with the life tally (B15LifeGainedThisTurn) of
// each opponent; the spell is a drain, and with its one target gone it does
// nothing at all, so no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8fedb0e-27d5-4fae-8058-3e9666a9661d",
		Name:         "Needlebite Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent gained life this turn, pay {B}",
			ManaCost:  "{B}",
			Condition: effects.Old495AnOpponentGainedLifeThisTurn,
		}},
		OnResolve: effects.ResolveSteps(effects.TargetPlayerLosesLife{N: 5}, effects.GainLife{Amount: 5}),
	})
}
