package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hornet Harasser — Creature — Goblin Shaman {2}{B}{B}, 2/2 (Lorwyn):
//
//	"When this creature dies, target creature gets -2/-2 until end of turn."
//
// Festering Goblin's dies trigger at -2/-2, over any creature. The target is
// chosen as the trigger goes on the stack (CR 603.3d); one that is gone or no
// longer legal on resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	t := effects.WhenThisDies("Hornet Harasser — target creature gets -2/-2 until end of turn",
		effects.BoostTargetUntilEOT("Hornet Harasser — -2/-2 until end of turn", -2, -2))
	t.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "5c0c3f08-4f65-4b04-8bfc-1532160f3e28",
		Name:         "Hornet Harasser",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
