package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rural Recruit — Creature — Human Peasant {3}{G}, 1/1 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 When this creature enters, create a 3/1 green Boar creature token."
//
// Training is the shared Training() trigger (training.go, CR 702.149a); the
// Boar is a plain token row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fda06d9e-d17f-4bb6-916b-06116e47c8e8",
		Name:         "Rural Recruit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			Training("Rural Recruit — training: put a +1/+1 counter on it"),
			WhenThisEnters("Rural Recruit — create a 3/1 green Boar creature token",
				Do(CreateToken{Template: TokenCard("3/1 green Boar"), N: 1})),
		},
	})
}
