package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unbender Tine — Artifact {2}{W}{U} (Alara Reborn, #5586):
//
//	"{T}: Untap another target permanent."
//
// Formidable Speaker's untap row without its mana: "another" is object
// identity (CR 109.1), so the Tine can't target itself, and a target that
// left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdc0a1a5-d9e4-4923-b472-2fa78212167f",
		Name:         "Unbender Tine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap another target permanent.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetPermanent("another target permanent", effects.Permanent())),
			Effect:  effects.B22UntapEachLegalTarget,
		}},
	})
}
