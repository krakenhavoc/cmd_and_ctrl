package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savageborn Hydra — Creature — Hydra {X}{R}{G}, 0/0 (Dragon's Maze, S82 OLD
// b405, #5564):
//
//	"Double strike
//	 This creature enters with X +1/+1 counters on it.
//	 {1}{R/G}: Put a +1/+1 counter on this creature. Activate only as a
//	 sorcery."
//
// Endless One's entry (X is announced with the cast, CR 601.2b, and the
// counters are part of the entry, CR 614.1c) plus Intrepid Tenderfoot's
// sorcery-speed counter on its own source (CR 602.5d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "fa8e5291-1a50-4c05-820c-8ae756364d33",
		Name:                       "Savageborn Hydra",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"double strike"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{R/G}: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
			Purpose:      game.Purpose{SourceCounters: 1},
			Cost:         effects.ManaCost("{1}{R/G}"),
			SorcerySpeed: true,
			Effect:       effects.PlusOneCountersOnThis(1),
		}},
	})
}
