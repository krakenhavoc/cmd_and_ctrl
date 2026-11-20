package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peace Strider — Artifact Creature — Construct {4}, 3/3 (Mirrodin Besieged,
// #5576):
//
//	"When this creature enters, you gain 3 life."
//
// Angel of Mercy's enters trigger: it uses the stack, so it can be
// answered, and the life goes to the trigger's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0e96742-77c0-4de5-84b0-0d59a8bee3e0",
		Name:         "Peace Strider",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Peace Strider — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
