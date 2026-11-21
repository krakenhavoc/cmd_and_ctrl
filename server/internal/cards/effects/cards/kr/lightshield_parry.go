package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightshield Parry — Instant {W} (Aetherdrift, #4464):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Artful Maneuver's pump without rebound, with cycling (CR 702.29a). One
// target, so a creature gone by resolution counters the spell
// (CR 608.2b); the pump ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecef9e3d-dce8-405f-bba3-14732444258f",
		Name:         "Lightshield Parry",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BoostTheTargetUntilEOT(2, 2, "Lightshield Parry — +2/+2"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
