package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mnemonic Sphere — Artifact {1}{U} (Neon Dynasty, #5493):
//
//	"{1}{U}, Sacrifice this artifact: Draw two cards.
//	 Channel — {U}, Discard this card: Draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53bfffae-b99e-44bd-8954-6cf236f93f26",
		Name:         "Mnemonic Sphere",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{U}, Sacrifice this artifact: Draw two cards.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{U}"), SacrificeThis()),
				Effect:  neoB07DrawsCards(2),
			},
			{
				Label:   "Channel — {U}, Discard this card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{U}"), DiscardThis()),
				Zones:   []game.ZoneKind{game.ZoneHand},
				Effect:  neoB07DrawsCards(1),
			},
		},
	})
}
