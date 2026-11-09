package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merrow Skyswimmer — Creature — Merfolk Soldier {3}{W/U}{W/U} (Lorwyn
// Eclipsed, #4458):
//
//	"Convoke
//	 Flying, vigilance
//	 When this creature enters, create a 1/1 white and blue Merfolk
//	 creature token."
func init() {
	Register(Spec{
		OracleID:        "d709f396-021f-49eb-adb5-374aa790eaae",
		Name:            "Merrow Skyswimmer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		TapCost:         Convoke(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Merrow Skyswimmer — create a 1/1 white and blue Merfolk creature token",
				Do(CreateToken{Template: TokenCard("1/1 white and blue Merfolk"), N: 1})),
		},
	})
}
