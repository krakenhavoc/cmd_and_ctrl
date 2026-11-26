package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Figment — Creature — Illusion {1}{U} 1/1 (Zendikar, S82 OLD batch
// 491, #5583):
//
//	"Kicker {3}
//	 If this creature was kicked, it enters with two +1/+1 counters on it.
//	 This creature can't be blocked."
//
// The counters are the entry replacement that reads the cast (CountersPerKick,
// as Shalai's Acolyte does); can't be blocked is the self restriction River
// Sneak has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "15fee0bd-ede2-4f65-a9a1-2da4623679d5",
		Name:                       "Aether Figment",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
