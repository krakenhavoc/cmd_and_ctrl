package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Crusher — Creature — Beast {3}{R}{R}, 4/4 (Nemesis, #5618):
//
//	"{R}: This creature gets +1/-1 until end of turn."
//
// Flowstone Mauler's pump without trample: the shared
// thisGetsUntilEndOfTurn body at +1/-1 for {R}, ending in the cleanup
// step (CR 514.2). A fourth activation leaves it 8/0, and it dies to the
// state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc18fcfc-bfe8-41d8-9037-5ff05b2a9c44",
		Name:         "Flowstone Crusher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/-1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Flowstone Crusher — +1/-1 until end of turn"),
		}},
	})
}
