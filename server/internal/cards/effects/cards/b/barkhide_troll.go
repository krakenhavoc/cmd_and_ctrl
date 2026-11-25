package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barkhide Troll — Creature — Troll {G}{G}, 2/2 (Core Set 2020, #5518):
//
//	"This creature enters with a +1/+1 counter on it.
//	 {1}, Remove a +1/+1 counter from this creature: This creature gains
//	 hexproof until end of turn."
//
// Monoskelion's entry counter (a self-replacement, CR 614.1c) and removal cost
// (paid whole before anything resolves, so a response cannot spend it twice);
// the hexproof is granted to this object only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7d231eb-255d-4505-89a9-0c6824622209",
		Name:         "Barkhide Troll",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove a +1/+1 counter from this creature: This creature gains hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("hexproof", "Barkhide Troll — hexproof until end of turn"),
		}},
	})
}
