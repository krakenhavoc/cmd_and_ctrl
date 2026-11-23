package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Astral Steel — Instant {2}{W} (Scourge):
//
//	"Target creature gets +1/+2 until end of turn.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Giant Growth's body for +1/+2, plus the storm keyword (CR 702.40a). Each
// copy is its own pump and may be pointed at a different creature (CR
// 707.10c); copies left on the same creature add up.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef8ac291-3789-4147-a11f-c62fd8566dbe",
		Name:         "Astral Steel",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 2)),
		Targets:      effects.TargetCreature("target creature"),
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve:    effects.BoostTheTargetUntilEOT(1, 2, "Astral Steel — +1/+2 until end of turn"),
	})
}
