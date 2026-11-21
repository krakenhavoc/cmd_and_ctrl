package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Scuta — Creature — Phyrexian Zombie {3}{B}, 3/3 (Planeshift,
// #5782):
//
//	"Kicker—Pay 3 life. (You may pay 3 life in addition to any other
//	 costs as you cast this spell.)
//	 If this creature was kicked, it enters with two +1/+1 counters on
//	 it."
//
// A kicker that pays a fixed amount of life (KickerPayLife, ADR 0100's
// 2026-11-20 amendment), offered only while the caster's life total is
// at least 3 (CR 119.4). The counters are read off the cast, as Gnarlid
// Colony's are: a single kicker is paid at most once, so a kicked Scuta
// enters as a 5/5.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "af2845cf-8770-4efa-8a83-bf9eb4122b78",
		Name:                       "Phyrexian Scuta",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.KickerPayLife(3)},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
	})
}
