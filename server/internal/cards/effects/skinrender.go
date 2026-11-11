package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skinrender — Creature — Phyrexian Zombie {2}{B}{B}, 3/3 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"When this creature enters, put three -1/-1 counters on target
//	 creature."
//
// A required target, any creature (it may be Skinrender itself when it is
// the only one). Three counters go on together (cmdr2B25CountersOnTarget);
// a target gone by resolution gets none (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c5c225a6-6a5e-4905-9412-dad15ba645b4",
		Name:         "Skinrender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Skinrender — put three -1/-1 counters on target creature", cmdr2B25CountersOnTarget(3)),
				Harmful(TargetCreature("target creature")),
			),
		},
	})
}
