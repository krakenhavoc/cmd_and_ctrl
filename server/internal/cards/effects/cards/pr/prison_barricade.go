package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prison Barricade — Creature — Wall {1}{W}, 1/3 (Invasion, S82 OLD #5616):
//
//	"Defender (This creature can't attack.)
//	 Kicker {1}{W} (You may pay an additional {1}{W} as you cast this spell.)
//	 If this creature was kicked, it enters with a +1/+1 counter on it and
//	 with 'This creature can attack as though it didn't have defender.'"
//
// The counter is the kicked entry (CR 614.1c). The permission is Skyclave
// Sentinel's can-attack-as-though static, held only while the permanent's own
// kicker record says it was kicked (game.CardKickedTimes, CR 400.7d), so a
// Barricade that was never cast has none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "8531b38a-8cc9-452b-8373-3febc733a95b",
		Name:                       "Prison Barricade",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"defender"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{1}{W}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 1)},
		Static: []game.StaticAbility{
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"If this creature was kicked, it can attack as though it didn't have defender.",
				func(_ *game.Game, source *game.Card) bool {
					return source != nil && game.CardKickedTimes(*source) > 0
				}),
		},
	})
}
