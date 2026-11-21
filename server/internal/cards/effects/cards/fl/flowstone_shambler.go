package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Shambler — Creature — Beast {2}{R}, 2/2 (Stronghold, #5627):
//
//	"{R}: This creature gets +1/-1 until end of turn."
//
// Breathstealer's +1/-1 at {R}, with the shared thisGetsUntilEndOfTurn
// body. A second activation leaves it 4/0, and it dies to the state-based
// check (CR 704.5f); the boosts end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e72ae9ab-7405-4ea7-b075-42d0ddc352db",
		Name:         "Flowstone Shambler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Flowstone Shambler — +1/-1 until end of turn"),
		}},
	})
}
