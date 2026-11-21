package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	t := effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))
	effects.Register(effects.Spec{
		OracleID:     "7b4aa101-3c8a-44b3-9d87-91a4fe4fbae4",
		Name:         "Ruinous Path",
		Completeness: effects.CompletenessFull,
		Targets:      t,
		AlternativeCosts: []game.AlternativeCost{
			effects.Awaken(4, "{5}{B}{B}", t),
		},
		OnResolve: effects.AwakenAfter(4, effects.DestroyAwakenSpellTarget),
	})
}
