package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necromancer's Assistant — Creature — Zombie {2}{B}, 3/1 (Magic 2015,
// #5556):
//
//	"When this creature enters, mill three cards."
//
// Aftermath Analyst's enters trigger: the controller mills three as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f72f111-450a-4da2-81fd-cbc4f9c1ef9d",
		Name:         "Necromancer's Assistant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Necromancer's Assistant — mill three cards", Do(MillCards{N: 3})),
		},
	})
}
