package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleshgrafter — Creature — Human Warrior {2}{B}, 2/2 (Fifth Dawn, #5604):
//
//	"Discard an artifact card: This creature gets +2/+2 until end of
//	 turn."
//
// Atog's +2/+2 paid with Artificer's Intuition's "discard an artifact card"
// cost instead of a sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04b1fb9e-4aef-46d9-b163-f9beff198bad",
		Name:         "Fleshgrafter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard an artifact card: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.DiscardCardsMatching(1, "an artifact card", effects.MatchArtifact),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 2, "Fleshgrafter — +2/+2 until end of turn"),
		}},
	})
}
