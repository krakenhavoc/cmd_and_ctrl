package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Informant — Creature — Human Rogue {2}{U}, 1/4 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, surveil 2."
//
// Sterling Hound's enter trigger: surveil 2 on the stack (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa66d63b-dfb7-42b8-8249-4d60e89058ac",
		Name:         "Dimir Informant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dimir Informant — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
