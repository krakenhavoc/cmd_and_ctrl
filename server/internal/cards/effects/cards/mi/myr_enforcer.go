package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Enforcer — Artifact Creature — Myr {7}, 4/4:
//
//	"Affinity for artifacts (This spell costs {1} less to cast for
//	 each artifact you control.)"
//
// #746: the whole card is its own cost modifier (CR 702.41a). The
// body is printed data the importer already carries.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d8e1054-654f-42e8-8c29-12c3cf13a3eb",
		Name:         "Myr Enforcer",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for artifacts", effects.Artifact()),
		},
	})
}
