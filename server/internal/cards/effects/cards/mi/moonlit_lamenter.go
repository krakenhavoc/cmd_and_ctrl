package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonlit Lamenter — Creature — Treefolk Cleric {2}{W}, 2/5 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with a -1/-1 counter on it.
//	 {1}{W}, Remove a counter from this creature: Draw a card. Activate
//	 only as a sorcery."
//
// Burdened Stoneback's entering counter and sorcery-speed removal cost,
// Loch Mare's draw. "A counter" prints no kind: the activator names which
// kind to remove (RemoveCountersFromThis("", 1)), so a +1/+1 counter it
// picked up pays too. The cost is paid at announce, so a response cannot
// spend the same counter twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce009bd4-5ffb-46a4-a9dd-b9c060926467",
		Name:         "Moonlit Lamenter",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMinusOne, 1, "Moonlit Lamenter: enters with a -1/-1 counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, Remove a counter from this creature: Draw a card. Activate only as a sorcery.",
			Purpose:      game.Purpose{Draws: 1},
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.RemoveCountersFromThis("", 1)),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
