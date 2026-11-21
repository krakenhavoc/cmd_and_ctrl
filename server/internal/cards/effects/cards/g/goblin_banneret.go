package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Banneret — Creature — Goblin Soldier {R}, 1/1 (Guilds of Ravnica, #5525):
//
//	"Mentor
//	 {1}{R}: This creature gets +2/+0 until end of turn."
//
// Mentor is the shared trigger (Blade Instructor's); the pump is Hematite
// Golem's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2d9554e-70c0-4213-ba5e-f0954ad32854",
		Name:         "Goblin Banneret",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Mentor("Goblin Banneret — mentor")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Goblin Banneret — +2/+0 until end of turn"),
		}},
	})
}
