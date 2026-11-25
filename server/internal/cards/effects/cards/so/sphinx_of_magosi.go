package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of Magosi — Creature — Sphinx {3}{U}{U}{U}, 6/6 (Conspiracy: Take
// the Crown, S82 older sets batch 272, #5540):
//
//	"Flying
//	 {2}{U}: Draw a card, then put a +1/+1 counter on this creature."
//
// Two steps in printed order (CR 608.2c), the draw first and then the
// counter on the Sphinx (a Sphinx that has left gets none).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c9cf20c-51e4-4a87-b9f8-231b6baf5d6b",
		Name:            "Sphinx of Magosi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Draw a card, then put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Draws: 1},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
