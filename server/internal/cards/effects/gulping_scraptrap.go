package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gulping Scraptrap — Creature — Phyrexian Horror {4}{B}, 4/4 (Phyrexia: All
// Will Be One, S82):
//
//	"When this creature enters or dies, proliferate."
//
// One printed trigger over two events (Mogg War Marshal's shape), so one
// registered trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d9e48750-d930-4095-abf0-dec2ac53be81",
		Name:         "Gulping Scraptrap",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Gulping Scraptrap — proliferate", Do(Proliferate{})),
		},
	})
}
