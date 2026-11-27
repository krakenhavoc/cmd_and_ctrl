package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Squadron (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Faerie {U}, 1/1:
//	"Kicker {3}{U}
//	If this creature was kicked, it enters with two +1/+1 counters on it and with flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "109f7a7d-dc08-4c8f-8065-93c620a9194e",
		Name:                       "Faerie Squadron",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}{U}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{effects.B16GrantKeywords(effects.GapB53ThisWasKicked, "flying")},
	})
}
