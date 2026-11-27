package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Split-Tail Miko — Creature — Fox Cleric {1}{W}, 1/1 (Betrayers of
// Kamigawa, #5602):
//
//	"{W}, {T}: Prevent the next 2 damage that would be dealt to any target
//	 this turn."
//
// Aven Redeemer's two-point charged shield (PreventNextDamageToTheTarget,
// CR 615.7) on the one target, behind a mana-and-tap cost. A target gone by
// resolution gets no shield (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82732de2-b98d-4534-967f-2cf4c9e6bfbe",
		Name:         "Split-Tail Miko",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Split-Tail Miko", 2),
		}},
	})
}
