package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge of Skola Vale — Creature — Hydra {2}{G}, 0/0 (Born of the Gods, S82
// OLD b379, #5559):
//
//	"Trample
//	 This creature enters with two +1/+1 counters on it.
//	 {T}, Sacrifice another creature: Put a number of +1/+1 counters on this
//	 creature equal to the sacrificed creature's toughness."
//
// Mindless Automaton's enters-with replacement, and Temple of Aclazotz's
// sacrificed-toughness read (as the creature last existed, CR 608.2h) as the
// counter count (S82OldB379ScourgeGrow).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "725d7374-4ea8-4aef-ac09-8dcb5d8dcad1",
		Name:            "Scourge of Skola Vale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Scourge of Skola Vale: enters with two +1/+1 counters on it"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature: Put a number of +1/+1 counters on this creature equal to the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.S82OldB379ScourgeGrow,
		}},
	})
}
