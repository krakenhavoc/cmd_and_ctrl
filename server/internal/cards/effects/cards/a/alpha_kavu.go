package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpha Kavu — Creature — Kavu {2}{G}, 2/2 (Planeshift, S82 OLD #5615):
//
//	"{1}{G}: Target Kavu creature gets -1/+1 until end of turn."
//
// Brave-Kin Duo's targeted pump body with a Kavu clause: any Kavu on the
// battlefield, this one included; a target that left, or stopped being a
// Kavu, in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1a76cca-829a-4576-a560-f1af34b80b87",
		Name:         "Alpha Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Target Kavu creature gets -1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target Kavu creature", effects.OfSubtype("Kavu")),
			Effect:  effects.BoostTargetUntilEOT("Alpha Kavu — -1/+1", -1, 1),
		}},
	})
}
