package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coral Barrier — Creature — Wall {2}{U}, 1/3 (Magic 2015, #5556):
//
//	"Defender
//	 When this creature enters, create a 1/1 blue Squid creature token with
//	 islandwalk."
//
// Defender is the printed keyword; the Squid is the existing catalog token
// (Magic 2015 prints the same 1/1 blue Squid with islandwalk).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f03474ce-b687-489f-add3-a82d77e18355",
		Name:            "Coral Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Coral Barrier — create a 1/1 blue Squid creature token with islandwalk",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Squid with islandwalk"), N: 1})),
		},
	})
}
