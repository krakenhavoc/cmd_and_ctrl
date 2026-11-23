package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Skyswimmer — Creature — Merfolk Soldier {3}{W/U}{W/U} (Lorwyn
// Eclipsed, #4458):
//
//	"Convoke
//	 Flying, vigilance
//	 When this creature enters, create a 1/1 white and blue Merfolk
//	 creature token."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d709f396-021f-49eb-adb5-374aa790eaae",
		Name:            "Merrow Skyswimmer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		TapCost:         effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Merrow Skyswimmer — create a 1/1 white and blue Merfolk creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and blue Merfolk"), N: 1})),
		},
	})
}
