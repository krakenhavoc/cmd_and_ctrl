package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bfdc6bbc-7402-40ae-91be-658afca2f145",
		Name:            "Spire Golem",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Islands", And(Land(), Subtype("Island"))),
		},
	})
}
