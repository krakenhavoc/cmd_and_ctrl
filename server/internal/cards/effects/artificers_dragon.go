package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artificer's Dragon — Artifact Creature — Dragon {6}, 4/4:
//
//	"Flying
//	 {R}: Artifact creatures you control get +1/+0 until end of turn.
//	 Unearth {3}{R}{R}"
func init() {
	Register(Spec{
		OracleID:        "4284ff9f-4085-4874-abc7-95467cb759a5",
		Name:            "Artificer's Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:   "{R}: Artifact creatures you control get +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(Artifact(), Creature(), YouControl()),
						Power: 1,
						Label: "Artificer's Dragon — artifact creatures you control get +1/+0",
					}.Apply(NewContext(g, item))
				},
			},
			Unearth("{3}{R}{R}"),
		},
	})
}
