package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Experimental Aviator — Creature — Human Artificer {3}{U}{U}, 0/3
// (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, create two 1/1 colorless Thopter artifact
//	 creature tokens with flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47f221a4-b51a-46b9-a4fc-46bc8ec7ddd3",
		Name:            "Experimental Aviator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Experimental Aviator — create two Thopters",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
	})
}
