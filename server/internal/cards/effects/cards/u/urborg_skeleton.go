package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urborg Skeleton (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Skeleton {B}, 0/1:
//	"Kicker {3}
//	{B}: Regenerate this creature.
//	If this creature was kicked, it enters with a +1/+1 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "712d7eb4-4044-4852-a5da-57635d3086b0",
		Name:                       "Urborg Skeleton",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
		Activated:                  []effects.ActivatedAbility{effects.RegenerateThisRow("{B}")},
	})
}
