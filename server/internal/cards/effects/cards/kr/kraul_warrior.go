package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraul Warrior — Creature — Insect Warrior {1}{G}, 2/2 (Ravnica: Clue
// Edition, Guilds of Ravnica):
//
//	"{5}{G}: This creature gets +3/+3 until end of turn."
//
// The shared self-pump body, ending in the cleanup step (CR 514.2); it can
// be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b05d000d-39b4-4db5-bd32-9c77a7434127",
		Name:         "Kraul Warrior",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: This creature gets +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(3, 3, "Kraul Warrior — +3/+3 until end of turn"),
		}},
	})
}
