package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Marauder — Artifact Creature — Phyrexian Construct {X}, 0/0
// (Visions, #5631):
//
//	"This creature enters with X +1/+1 counters on it.
//	 This creature can't block.
//	 This creature can't attack unless you pay {1} for each +1/+1 counter on
//	 it."
//
// Hangarback Walker's X counters, Aesthir Glider's "can't block" and Myr
// Prototype's attack tax (#6191), the tax read as the attack is declared
// (CR 508.1h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "405674c5-607a-43d3-8b46-8741d2339f80",
		Name:                       "Phyrexian Marauder",
		Completeness:               effects.CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		XMatters:                   true,
		Static:                     []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		AttackTaxes:                []game.AttackTax{effects.CantAttackUnlessYouPayPerCounter("+1/+1", "This creature can't attack unless you pay {1} for each +1/+1 counter on it.")},
	})
}
