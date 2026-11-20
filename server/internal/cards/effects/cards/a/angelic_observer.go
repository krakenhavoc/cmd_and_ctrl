package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Observer — Creature — Angel Advisor {5}{W}, 3/3 (Streets of New
// Capenna, #5491):
//
//	"Affinity for Citizens (This spell costs {1} less to cast for each
//	 Citizen you control.)
//	 Flying"
//
// Allies at Last's shared affinity self cost modifier over the Citizens you
// control; it reduces only the generic part of the cost (CR 702.41a). Flying
// is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17df0fda-d726-4c9d-bf44-6f2e3a347117",
		Name:            "Angelic Observer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Citizens", effects.HasSubtype("Citizen")),
		},
	})
}
