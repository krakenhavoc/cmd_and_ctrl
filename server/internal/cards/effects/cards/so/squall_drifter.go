package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squall Drifter — Snow Creature — Elemental {1}{W}, 1/1 (Coldsnap):
//
//	"Flying
//	 {W}, {T}: Tap target creature."
//
// The {T} in the cost needs the Drifter free of summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63ec5617-0da6-4534-b502-34304b1ea02b",
		Name:            "Squall Drifter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.ThousandMoonsCrackshotTap,
		}},
	})
}
