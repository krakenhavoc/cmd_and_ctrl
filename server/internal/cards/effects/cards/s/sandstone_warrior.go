package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstone Warrior — Creature — Human Soldier Warrior {2}{R}{R}, 1/3
// (Tempest, S82 OLD batch 757, #5628):
//
//	"First strike
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Bellows Lizard's firebreathing at {R}, plus a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "120a0bcd-eca2-44eb-ad8d-5e4077f7e21c",
		Name:            "Sandstone Warrior",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Sandstone Warrior — +1/+0 until end of turn"),
		}},
	})
}
