package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resolute Reinforcements — Creature — Human Soldier {1}{W}, 1/1
// (Foundations, #4465):
//
//	"Flash
//	 When this creature enters, create a 1/1 white Soldier creature
//	 token."
//
// Flash is an engine keyword carried as printed data; the token is the
// shared "1/1 white Soldier" row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eaaed253-9b11-47e2-8624-0340ab2207f6",
		Name:            "Resolute Reinforcements",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Resolute Reinforcements — create a Soldier", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
