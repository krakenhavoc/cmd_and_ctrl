package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tymora's Invoker — Creature — Orc Rogue {1}{U}, 1/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Sleight of Hand — {8}: Draw two cards."
//
// An activated ability with a pure mana cost and no tap, so summoning sickness
// does not apply and it may be activated at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36652a44-f5bd-4bc0-abfd-fa3688853dfd",
		Name:         "Tymora's Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sleight of Hand — {8}: Draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{8}"),
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}
