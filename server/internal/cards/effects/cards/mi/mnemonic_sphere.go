package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mnemonic Sphere — Artifact {1}{U} (Neon Dynasty, #5493):
//
//	"{1}{U}, Sacrifice this artifact: Draw two cards.
//	 Channel — {U}, Discard this card: Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53bfffae-b99e-44bd-8954-6cf236f93f26",
		Name:         "Mnemonic Sphere",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}, Sacrifice this artifact: Draw two cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.SacrificeThis()),
				Effect:  effects.NeoB07DrawsCards(2),
			},
			{
				Label:   "Channel — {U}, Discard this card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Effect:  effects.NeoB07DrawsCards(1),
			},
		},
	})
}
