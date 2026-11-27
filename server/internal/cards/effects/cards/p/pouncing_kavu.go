package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pouncing Kavu (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Kavu {1}{R}, 1/1:
//	"Kicker {2}{R}
//	First strike
//	If this creature was kicked, it enters with two +1/+1 counters on it and with haste."
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "5f9a968d-cc23-46d3-b28f-3779dd89c531",
		Name:                       "Pouncing Kavu",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"first strike"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}{R}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{effects.B16GrantKeywords(effects.GapB53ThisWasKicked, "haste")},
	})
}
