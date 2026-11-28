package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Ski Patrol — Creature — Goblin {1}{R}, 1/1 (Ice Age, S82 OLD, #5636):
//
//	"{1}{R}: This creature gets +2/+0 and gains flying. Its controller sacrifices
//	 it at the beginning of the next end step. Activate only once and only if you
//	 control a snow Mountain."
//
// "Only once" is the ability's OnlyOnce bit, and the Mountain is read as a word of the card's
// text (CR 612.2).
//
// No simplification.
func init() {
	const label = "{1}{R}: This creature gets +2/+0 and gains flying. Its controller sacrifices it at the beginning of the next end step. Activate only once and only if you control a snow Mountain."
	effects.Register(effects.Spec{
		OracleID:     "e98396d0-6952-4755-992f-813f565b9189",
		Name:         "Goblin Ski Patrol",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:      effects.ManaCost("{1}{R}"),
			Condition: effects.OldB835SnowMountain(),
			OnlyOnce:  true,
			Effect:    effects.OldB835ThisGetsAndFliesThenSacrificeAtEnd,
		}},
	})
}
