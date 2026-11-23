package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soltari Crusader — Creature — Soltari Knight {2}{W}, 2/1
// (Tempest, #5628):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 {1}{W}: This creature gets +1/+0 until end of turn."
//
// Dauthi Mercenary's shape in white: shadow is a printed keyword, and
// the pump is the shared thisGetsUntilEndOfTurn body at {1}{W}, ending
// in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "39e01990-9928-440c-a5c3-82aea29e2cd0",
		Name:            "Soltari Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Soltari Crusader — +1/+0"),
		}},
	})
}
