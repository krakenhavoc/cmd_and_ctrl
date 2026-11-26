package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of the Chimes — Creature — Sphinx {4}{U}{U}, 5/6 (Return to Ravnica,
// S82 OLD):
//
//	"Flying
//	 Discard two nonland cards with the same name: Draw four cards."
//
// The cost is the engine's same-name discard (DiscardNWithTheSameName, #6264,
// CR 201.2a): the activation refuses two cards with different names or a
// land. It costs no mana and no tap, so it can be activated any number of
// times with enough pairs in hand. The four cards are drawn as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "690df8d2-9538-4c51-b730-cee09ad0ef1c",
		Name:            "Sphinx of the Chimes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard two nonland cards with the same name: Draw four cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 4},
			Cost: effects.DiscardNWithTheSameName(2, "two nonland cards with the same name",
				func(c game.Card) bool { return !c.IsLand() }),
			Effect: effects.Do(effects.DrawCards{N: 4}),
		}},
	})
}
