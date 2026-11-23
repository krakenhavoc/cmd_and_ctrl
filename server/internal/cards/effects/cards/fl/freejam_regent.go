package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freejam Regent — Creature — Cat Dragon {4}{R}{R}, 4/4:
//
//	"Improvise ... Flying
//	 {1}{R}: This creature gets +2/+0 until end of turn."
//
// Improvise is Spec.TapCost; the pump is Canyon Drake's shape without the
// discard. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87a72c3d-e787-4dd8-b934-4fff4573af6c",
		Name:            "Freejam Regent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         effects.Improvise(),
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Freejam Regent — +2/+0"),
		}},
	})
}
