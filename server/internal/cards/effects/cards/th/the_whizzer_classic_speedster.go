package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Whizzer, Classic Speedster — Legendary Creature — Human Hero
// {3}{R}, 3/3 (Marvel Super Heroes Commander, #4455): "First strike,
// haste". A keyword-only body, the same shape as Glamorous Grapplers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa86bce1-9cde-44d8-9455-24502221f31d",
		Name:            "The Whizzer, Classic Speedster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
