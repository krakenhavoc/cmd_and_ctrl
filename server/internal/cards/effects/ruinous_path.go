package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruinous Path — Sorcery {1}{B}{B}:
//
//	"Destroy target creature or planeswalker.
//	 Awaken 4—{5}{B}{B}"
//
// ADR 0135 §3 (#2411): the destroy, then the awaken land (CR 702.113a),
// each target checked on its own (CR 608.2b).
//
// No simplifications.
func init() {
	t := TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))
	Register(Spec{
		OracleID:     "7b4aa101-3c8a-44b3-9d87-91a4fe4fbae4",
		Name:         "Ruinous Path",
		Completeness: CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			Awaken(4, "{5}{B}{B}", t),
		},
		OnResolve: AwakenAfter(4, destroyAwakenSpellTarget),
	})
}
