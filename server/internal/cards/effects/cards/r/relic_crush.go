package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Crush — Instant {4}{G} (Commander 2017, #5533):
//
//	"Destroy target artifact or enchantment and up to one other target
//	 artifact or enchantment."
//
// Two target clauses read positionally, the second "up to one" and Distinct
// so it cannot be the first (Bionic Blow's shape). Every target still legal
// as the spell resolves is destroyed; if the first has gone the second is
// still destroyed (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92381398-94f6-471f-874a-b5894e6f9487",
		Name:         "Relic Crush",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			effects.Distinct(effects.TargetPermanent("up to one other target artifact or enchantment",
				effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)),
		),
		OnResolve: effects.DestroyEachLegalTarget,
	})
}
