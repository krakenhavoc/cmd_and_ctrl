package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Amateur Hero — {2}{W} Creature — Human Hero, 3/3 (Marvel's Spider-Man
// Eternal, #4460):
//
//	"When this creature enters, you gain 2 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc8ee6e1-a2e1-4033-a164-7968fc79dd27",
		Name:         "Amateur Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Amateur Hero — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
