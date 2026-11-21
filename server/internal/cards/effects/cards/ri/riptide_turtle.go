package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riptide Turtle — Creature — Turtle {1}{U}, 0/5 (Theros Beyond Death, #5514):
//
//	"Flash
//	 Defender"
//
// Flash and defender are the printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb7e076c-a695-4c79-9f20-498370c81380",
		Name:            "Riptide Turtle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "defender"},
	})
}
