package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valorous Steed — Creature — Unicorn {4}{W}, 3/3:
//
//	"Vigilance
//	 When this creature enters, create a 2/2 white Knight creature token
//	 with vigilance."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b5f30a6-b0fb-4dcb-85d5-1db1bba73737",
		Name:            "Valorous Steed",
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Valorous Steed — create a 2/2 white Knight token with vigilance",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
	})
}
