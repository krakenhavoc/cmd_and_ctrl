package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wary Watchdog — Creature — Dog {1}{G}, 3/1 (Duskmourn: House of
// Horror, #4466):
//
//	"When this creature enters or dies, surveil 1."
//
// Wary Thespian's enters-or-dies trigger: one ability with two
// conditions, so it surveils once as it arrives and once more as it
// dies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e063f60-876a-48fe-8bc6-9888180176d1",
		Name:         "Wary Watchdog",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Wary Watchdog — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
