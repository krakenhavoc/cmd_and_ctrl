package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S.H.I.E.L.D. Deployment Drone — {2}{U} Artifact Creature — Robot, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 When this creature enters, create a 1/1 white Soldier creature token."
//
// Black Panther, Vanguard's Soldier token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1669020-8f81-471a-9083-7e9ce3e8f387",
		Name:            "S.H.I.E.L.D. Deployment Drone",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("S.H.I.E.L.D. Deployment Drone — create a 1/1 white Soldier", Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
