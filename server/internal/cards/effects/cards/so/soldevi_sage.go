package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soldevi Sage — Creature — Human Wizard {1}{U}, 1/1 (Alliances, #5634):
//
//	"{T}, Sacrifice two lands: Draw three cards, then discard one of them."
//
// The cost taps the Sage and sacrifices two lands (SacrificeN, Sai's shape).
// "One of them" is one of the three cards just drawn, not a card that was
// already in hand, so the discard is a pick over the drawn cards only
// (OldB814DrawThreeThenDiscardOneOfThem).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f612df3-53b6-4317-9d63-1f903ee3f0c4",
		Name:         "Soldevi Sage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice two lands: Draw three cards, then discard one of them.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(2, "two lands", effects.Land())),
			Effect:  effects.OldB814DrawThreeThenDiscardOneOfThem,
		}},
	})
}
