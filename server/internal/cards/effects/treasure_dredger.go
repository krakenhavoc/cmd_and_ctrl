package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treasure Dredger — Creature — Human Rogue {1}{B}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"{1}, {T}, Pay 1 life: Create a Treasure token."
//
// Pictures of Spider-Man's Treasure row with Fountainport's life payment
// in the cost. The {T} waits out summoning sickness (CR 302.6). An
// untargeted instant-speed row that answers nothing, so it declares
// AnswerValue (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a691e96-fb47-4c23-a063-3b287c51147b",
		Name:         "Treasure Dredger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Pay 1 life: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost(), PayLife(1)),
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
