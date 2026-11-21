package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakdos Trumpeter — Creature — Human Shaman {1}{B}, 1/3 (Ravnica: Clue Edition, #5929):
//
//	"Menace
//	 {3}{R}: This creature gets +2/+0 until end of turn."
//
// Menace is a printed keyword; the pump is Fangblade Brigand's self pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2047a613-809d-4a47-9c78-f58a380ba18d",
		Name:            "Rakdos Trumpeter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{R}"),
			Effect:  effects.ThisCreatureUntilEOT("Rakdos Trumpeter — +2/+0", 2, 0),
		}},
	})
}
