package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Branchrider — Creature — Lizard Warrior {R}, 1/1 (Dominaria
// United, #5487):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this
//	 spell.)
//	 Haste
//	 If this creature was kicked, it enters with two +1/+1 counters on it.
//	 {2}{R}: This creature gets +2/+0 until end of turn."
//
// Vodalian Serpent's kicked entry counters plus Canyon Drake's pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "50fcb431-626c-4db5-aa2c-445f3b81e7fa",
		Name:                       "Viashino Branchrider",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"haste"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{2}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Viashino Branchrider — +2/+0"),
		}},
	})
}
