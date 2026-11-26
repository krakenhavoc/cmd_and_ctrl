package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shriekhorn — Artifact {1} (Mirrodin Besieged, S82 OLD batch 464, #5576):
//
//	"This artifact enters with three charge counters on it.
//	 {T}, Remove a charge counter from this artifact: Target player mills
//	 two cards."
//
// An entry self-replacement for the counters (CR 614.1c) and Millstone's
// mill, paid with a tap and a charge counter from this artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "571c1923-790f-48b7-86d2-70c3e79fdba3",
		Name:         "Shriekhorn",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterCharge, 3, "Shriekhorn — enters with three charge counters on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove a charge counter from this artifact: Target player mills two cards.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(game.CounterCharge, 1)),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.B31MillChosenPlayer(2),
		}},
	})
}
