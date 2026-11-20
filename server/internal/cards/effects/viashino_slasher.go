package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Viashino Slasher — Creature — Lizard Warrior {1}{R}, 1/2 (Ravnica: City of
// Guilds, #5600):
//
//	"{R}: This creature gets +1/-1 until end of turn."
//
// Flowstone Shambler's +1/-1 row: any number of activations, each ending
// at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63908256-a8a4-494f-b5b1-4b92c064af1d",
		Name:         "Viashino Slasher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, -1, "Viashino Slasher — +1/-1 until end of turn"),
		}},
	})
}
