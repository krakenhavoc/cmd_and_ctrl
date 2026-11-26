package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grimdancer — Creature — Nightmare {1}{B}{B}, 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"This creature enters with your choice of two different counters on it
//	 from among menace, deathtouch, and lifelink."
//
// One question offers the three pairs, which is every way to choose two
// different counters from three. The printed deathtouch is the creature's
// own; a deathtouch counter is a second source of it and changes nothing
// until the printed one is lost (CR 122.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4afd88fa-f8bb-46d4-9ba7-09419e6f858c",
		Name:            "Grimdancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		AsEnters: effects.S82OldB113EntersWithChoiceOfCounters("Grimdancer",
			[]string{game.CounterMenace, game.CounterDeathtouch},
			[]string{game.CounterMenace, game.CounterLifelink},
			[]string{game.CounterDeathtouch, game.CounterLifelink}),
	})
}
