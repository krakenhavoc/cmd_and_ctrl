package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conch Horn — Artifact {2} (Fallen Empires, S82 OLD b855, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: Draw two cards, then put a card from your hand
//	 on top of your library."
//
// Survivor of the Unseen's draw and put-back: the card put back may be one just drawn.
// The sacrifice is a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b019cc20-d32d-4cdf-80cc-eafb3f78db88",
		Name:         "Conch Horn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Draw two cards, then put a card from your hand on top of your library.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.OldB559DrawTwoThenPutOneBack,
		}},
	})
}
