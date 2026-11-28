package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duskwalker (Invasion, S82 OLD batch 648, #5616):
//
//	Creature — Human Minion {B}, 1/1:
//	"Kicker {3}{B}
//	If this creature was kicked, it enters with two +1/+1 counters on it and with fear."
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "689f8696-dd13-4e1e-bdbc-3794c524f423",
		Name:                       "Duskwalker",
		Completeness:               effects.CompletenessFull,
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}{B}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static:                     []game.StaticAbility{effects.B16GrantKeywords(effects.GapB53ThisWasKicked, "fear")},
	})
}
