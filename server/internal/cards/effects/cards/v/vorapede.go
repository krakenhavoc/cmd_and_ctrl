package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vorapede — Creature — Insect {2}{G}{G}{G}, 5/4 (Dark Ascension, #5571):
//
//	"Vigilance, trample
//	 Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with
//	 a +1/+1 counter on it.)"
//
// Vigilance, trample and undying are PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd44986c-cf37-4f6e-b8d6-1795ec794b5f",
		Name:            "Vorapede",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", game.KeywordUndying},
	})
}
