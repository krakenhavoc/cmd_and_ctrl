package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Fire Fiend — Creature — Goblin Berserker {3}{R}, 1/1 (Ravnica: City of
// Guilds, S82 OLD batch 575, #5600):
//
//	"Haste
//	 This creature must be blocked if able.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Haste is the printed keyword; the second line is the engine's must-be-
// blocked requirement (CR 509.1c); the ability is Bellows Lizard's firebreathing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61f84dbb-e9bb-45b1-a30e-c631be6ad25c",
		Name:            "Goblin Fire Fiend",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.MustBeBlocked()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Goblin Fire Fiend — +1/+0 until end of turn"),
		}},
	})
}
