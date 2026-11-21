package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spire Golem — Artifact Creature — Golem {6}, 2/4:
//
//	"Affinity for Islands (This spell costs {1} less to cast for each
//	 Island you control.)
//	 Flying"
//
// Icebreaker Kraken's affinity shape: a self cost modifier counting the
// Islands (lands with the Island subtype, basic or not) its caster
// controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bfdc6bbc-7402-40ae-91be-658afca2f145",
		Name:            "Spire Golem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Islands", effects.And(effects.Land(), effects.Subtype("Island"))),
		},
	})
}
