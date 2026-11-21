package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aspiring Aeronaut — Creature — Human Artificer {3}{U}, 1/2 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 When this creature enters, create a 1/1 colorless Thopter artifact creature token with flying."
//
// Flying, and Nimble Thopterist's entry token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2e31c0f-8914-4a18-bc05-44efb7e0f18c",
		Name:            "Aspiring Aeronaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aspiring Aeronaut — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
