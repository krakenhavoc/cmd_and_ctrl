package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember-Eye Wolf — Creature — Wolf {1}{R}, 1/2
// (Shadows over Innistrad, #5543):
//
//	"Haste
//	 {1}{R}: This creature gets +2/+0 until end of turn."
//
// Hematite Golem's firebreathing at {1}{R} for +2/+0, with haste carried
// as printed data (CR 702.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0b21ab85-2405-4410-9624-fb3c7d11ba56",
		Name:            "Ember-Eye Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Ember-Eye Wolf — +2/+0 until end of turn"),
		}},
	})
}
