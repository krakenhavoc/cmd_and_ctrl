package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treasure Dredger — Creature — Human Rogue {1}{B}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"{1}, {T}, Pay 1 life: Create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// The life is part of the cost (CR 119.4), so it can't be paid at 0 life,
// and {T} is a creature's tap cost, so the Dredger can't activate the turn
// it arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a691e96-fb47-4c23-a063-3b287c51147b",
		Name:         "Treasure Dredger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Pay 1 life: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{1}"), TapCost(), PayLife(1)),
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}
