package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voidpouncer — Creature — Eldrazi {1}{R}, 3/1 (Modern Horizons 3):
//
//	"Devoid (This card has no color.)
//	 Kicker {2}{C} (You may pay an additional {2}{C} as you cast this
//	 spell.)
//	 If this creature was kicked, it enters with two +1/+1 counters and a
//	 trample counter on it and with haste."
//
// Gnarlid Colony's kicked entry counters (two +1/+1 counters and one
// trample counter, whose trample is the engine's keyword counter) and a
// layer 6 haste for a Voidpouncer that arrived by a kicked spell. The {C}
// of the kicker is colourless mana specifically.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb4ebe30-2501-48aa-879a-2402a9eb1be7",
		Name:            "Voidpouncer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{2}{C}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerKick(game.CounterPlusOne, 2),
			effects.CountersPerKick(game.CounterTrample, 1),
		},
		Static: []game.StaticAbility{effects.B16GrantKeywords(effects.GapB53ThisWasKicked, "haste")},
	})
}
