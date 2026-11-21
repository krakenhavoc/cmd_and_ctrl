package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "c5c225a6-6a5e-4905-9412-dad15ba645b4",
		Name:         "Skinrender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Skinrender — put three -1/-1 counters on target creature", effects.Cmdr2B25CountersOnTarget(3)),
				effects.Harmful(effects.TargetCreature("target creature")),
			),
		},
	})
}
