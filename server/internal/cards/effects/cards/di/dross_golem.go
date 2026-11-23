package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dross Golem — Artifact Creature — Golem {5} (Darksteel, #5605):
//
//	"Affinity for Swamps (This spell costs {1} less to cast for each
//	 Swamp you control.)
//	 Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)"
//
// Spire Golem's shape: affinity is the shared self cost modifier counting
// the Swamps its caster controls (CR 702.41a), and fear is a printed
// keyword (CR 702.36b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "336d51cf-dcb3-4257-bd6c-926820a2f9de",
		Name:            "Dross Golem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"fear"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Swamps", effects.And(effects.Land(), effects.Subtype("Swamp"))),
		},
	})
}
