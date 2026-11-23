package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furyborn Hellkite — Creature — Dragon {4}{R}{R}{R}, 6/6 (Magic 2012,
// #5573):
//
//	"Bloodthirst 6 (If an opponent was dealt damage this turn, this
//	 creature enters with six +1/+1 counters on it.)
//	 Flying"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); flying is
// a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c4aba38-7fbd-463e-8454-21037090c71e",
		Name:            "Furyborn Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.Bloodthirst(6, "Furyborn Hellkite")},
	})
}
