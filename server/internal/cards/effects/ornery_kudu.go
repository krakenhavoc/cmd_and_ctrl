package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ornery Kudu — Creature — Antelope {2}{G}, 3/4
// (Amonkhet, #5535):
//
//	"When this creature enters, put a -1/-1 counter on target creature you control."
//
// Channeler Initiate's enters trigger with one counter; the target may
// be this creature itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8abef87a-4a5c-412e-8137-827cee50bb8b",
		Name:         "Ornery Kudu",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ornery Kudu — put a -1/-1 counter on target creature you control", cmdr2B25CountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl()),
			),
		},
	})
}
