package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howling Giant — Creature — Giant Druid {5}{G}{G}, 5/5 (Core Set 2020, #5518):
//
//	"Reach (This creature can block creatures with flying.)
//	 When this creature enters, create two 2/2 green Wolf creature tokens."
//
// Tovolar's Huntmaster's enters trigger with a printed reach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c616242a-95f0-4cd8-99a1-9237cfea0e57",
		Name:            "Howling Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Tokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Howling Giant — create two 2/2 green Wolf creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 2})),
		},
	})
}
