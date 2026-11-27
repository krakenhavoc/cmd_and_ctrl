package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Executioner — Creature — Zombie {3}{B}, 2/2 (Avacyn Restored, S82
// OLD, #5570):
//
//	"When this creature dies, you may have target creature get -2/-2 until
//	 end of turn."
//
// Disease Carriers' dies trigger made a CR 603.5 "you may": the target is
// chosen as the ability goes on the stack, and the question is asked as it
// resolves. "No", or a target gone by then, does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2bfac84-0e2f-4f3e-8d87-a555c26093a6",
		Name:         "Undead Executioner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(effects.Targeting(
				effects.WhenThisDies("Undead Executioner — target creature gets -2/-2 until end of turn",
					effects.BoostTargetUntilEOT("Undead Executioner — -2/-2 until end of turn", -2, -2)),
				effects.TargetCreature("target creature"),
			), "Undead Executioner — have target creature get -2/-2 until end of turn?"), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
