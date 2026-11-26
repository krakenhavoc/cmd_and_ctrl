package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Etched Monstrosity — Artifact Creature — Phyrexian Golem {5}, 10/10
// (New Phyrexia):
//
//	"This creature enters with five -1/-1 counters on it.
//	 {W}{U}{B}{R}{G}, Remove five -1/-1 counters from this creature:
//	 Target player draws three cards."
//
// The five counters are a self-replacement on entry (CR 614.1c), so it is
// a 5/5 from the first moment. The cost removes all five or it cannot be
// paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0dd582a-ac35-4a46-beca-6304fdd2d358",
		Name:         "Etched Monstrosity",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMinusOne, 5, "Etched Monstrosity: enters with five -1/-1 counters on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, Remove five -1/-1 counters from this creature: Target player draws three cards.",
			Cost:    effects.Plus(effects.ManaCost("{W}{U}{B}{R}{G}"), effects.RemoveCountersFromThis(game.CounterMinusOne, 5)),
			Targets: effects.TargetPlayer("target player"),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3}),
			Effect:  effects.TheTargetPlayerDraws(3),
		}},
	})
}
