package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necrogen Censer — Artifact {3} (Scars of Mirrodin, S82 OLD batch 470,
// #5577):
//
//	"This artifact enters with two charge counters on it.
//	 {T}, Remove a charge counter from this artifact: Target player loses 2
//	 life."
//
// Two charge counters as it enters and one removed as the cost, so it can
// drain twice; a target that has left the game loses nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1feac23-8ba3-440c-828c-590f88008c31",
		Name:         "Necrogen Censer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterCharge, 2, "Necrogen Censer: enters with two charge counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a charge counter from this artifact: Target player loses 2 life.",
			Purpose: game.Purpose{},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.OldB470TargetPlayerLosesLife(2),
		}},
	})
}
