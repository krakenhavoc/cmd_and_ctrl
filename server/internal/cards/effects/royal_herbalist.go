package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Royal Herbalist — Creature — Human Cleric {W}, 1/1:
//
//	"{2}, Exile the top card of your library: You gain 1 life."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3baf498-3be7-4a22-8fed-806d8d8ac748",
		Name:         "Royal Herbalist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Exile the top card of your library: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), ExileTopOfLibrary(1)),
			Effect:  Do(GainLife{Amount: 1}),
		}},
	})
}
