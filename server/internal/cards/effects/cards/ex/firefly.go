package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firefly — Creature — Insect {3}{R}, 1/1 (Tempest, #5628):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Shivan Dragon's text on a 1/1: flying is a printed keyword, and the
// firebreathing is the shared thisGetsUntilEndOfTurn body at {R}, ending
// in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b151418-092f-4a43-8b21-3de64ea80cab",
		Name:            "Firefly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Firefly — +1/+0"),
		}},
	})
}
