package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lone Missionary — Creature — Kor Monk {1}{W}, 2/1 (Rise of the Eldrazi,
// #5581):
//
//	"When this creature enters, you gain 4 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "626f1dc3-3b14-4873-902f-ffec5487a97d",
		Name:         "Lone Missionary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lone Missionary — you gain 4 life", Do(GainLife{Amount: 4})),
		},
	})
}
