package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Two-Headed Dragon — {4}{R}{R} Creature — Dragon 4/4 (Archenemy):
//
//	"Flying
//	 Menace
//	 This creature can block an additional creature each combat.
//	 {1}{R}: This creature gets +2/+0 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a4fa3052-d1ec-4442-a832-15ba36c500ee",
		Name:            "Two-Headed Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Static:          []game.StaticAbility{effects.CanBlockAdditional(effects.SelfOnly, 1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Two-Headed Dragon — +2/+0"),
		}},
	})
}
