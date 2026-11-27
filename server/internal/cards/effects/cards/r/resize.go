package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resize — Instant {1}{G} (Coldsnap):
//
//	"Target creature gets +3/+3 until end of turn.
//	 Recover {1}{G}"
//
// Recover (CR 702.59a) is S82OldB564Recover.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c092ed4-4dbb-41ec-9e07-e59b9676802e",
		Name:         "Resize",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(3, 3, "Resize — +3/+3 until end of turn"),
		Triggered:    []game.TriggeredAbility{effects.S82OldB564Recover("Resize", "{1}{G}")},
	})
}
