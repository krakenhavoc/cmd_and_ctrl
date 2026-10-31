package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S.H.I.E.L.D. Helicarrier — {4} Artifact — Vehicle, 4/5 (Marvel Super
// Heroes, #4455):
//
//	"Flying
//	 When this Vehicle enters, create two 1/1 white Soldier creature tokens.
//	 Crew 6"
//
// The two tokens are one CreateToken with N: 2, so they enter together.
// Crew is Aethersphere Harvester's row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0f2335bf-6a13-4483-997a-fd19d30c4d48",
		Name:            "S.H.I.E.L.D. Helicarrier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("S.H.I.E.L.D. Helicarrier — create two 1/1 white Soldiers", Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 2})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 6",
				Cost:    CrewCost(6),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("S.H.I.E.L.D. Helicarrier"),
			},
		},
	})
}
