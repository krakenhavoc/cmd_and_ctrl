package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyclave Sentinel — Artifact Creature — Gargoyle {3}, 2/3 (Zendikar
// Rising, #5508):
//
//	"Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	 Flying, defender
//	 If this creature was kicked, it enters with two +1/+1 counters on
//	 it.
//	 As long as this creature has a +1/+1 counter on it, it can attack as
//	 though it didn't have defender."
//
// The kicked entry is Skyclave Shade's (CountersPerKick, CR 614.1c). The
// attack permission is Scuttlegator's static, read live off the counter
// map, so a counter from any source lets it attack and losing the last
// one takes the permission away.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3b7b166b-4329-4194-a9a7-f1fec7a92cec",
		Name:                       "Skyclave Sentinel",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying", "defender"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{4}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Static: []game.StaticAbility{
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"As long as this creature has a +1/+1 counter on it, it can attack as though it didn't have defender.",
				func(_ *game.Game, source *game.Card) bool {
					return source != nil && source.Counters[game.CounterPlusOne] > 0
				}),
		},
	})
}
