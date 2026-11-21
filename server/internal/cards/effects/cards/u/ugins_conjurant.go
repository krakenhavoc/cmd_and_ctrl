package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ugin's Conjurant — Creature — Spirit Monk {X}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 If damage would be dealt to this creature while it has a +1/+1 counter on it, prevent that damage and remove that many +1/+1 counters from this creature."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect. It
// removes "that many" counters — the damage, prevented or not, so damage
// that can't be prevented still removes them (CR 615.12) — or every
// counter it has when that is fewer, and all of the damage is still
// prevented (the ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "787055e6-5d65-458e-a56e-66ef0574d2ca",
		Name:                       "Ugin's Conjurant",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Replacements: []game.ReplacementEffect{
			effects.PreventDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				While: effects.WhileItHasAPlusOneCounter,
				Then:  effects.RemoveThatManyCountersFromThisBody,
				Label: "Ugin's Conjurant — prevent damage to it and remove that many +1/+1 counters",
			}),
		},
	})
}
