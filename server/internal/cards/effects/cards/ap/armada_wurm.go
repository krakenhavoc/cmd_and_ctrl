package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armada Wurm — Creature — Wurm {2}{G}{G}{W}{W}, 5/5 (Return to Ravnica,
// #5566):
//
//	"Trample
//	 When this creature enters, create a 5/5 green Wurm creature token with
//	 trample."
//
// The token is Hostility's and Worldspine Wurm's table entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "33598888-0367-4085-9415-f4e21da08354",
		Name:            "Armada Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Armada Wurm — create a 5/5 green Wurm creature token with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 green Wurm with trample"), N: 1})),
		},
	})
}
