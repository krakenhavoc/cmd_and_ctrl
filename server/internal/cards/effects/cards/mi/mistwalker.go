package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mistwalker — Creature — Shapeshifter {2}{U}, 1/4 (Kaldheim, #5505):
//
//	"Changeling (This card is every creature type.)
//	 Flying
//	 {1}{U}: This creature gets +1/-1 until end of turn."
//
// Changeling and flying ride PrintedKeywords; the pump is the shared
// thisGetsRow (Flowstone Crusher's +1/-1), activatable any number of
// times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ceb302df-d67e-4e5d-aca0-f53bcdea658d",
		Name:            "Mistwalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "flying"},
		Activated:       []effects.ActivatedAbility{effects.ThisGetsRow("{1}{U}", 1, -1, "Mistwalker")},
	})
}
