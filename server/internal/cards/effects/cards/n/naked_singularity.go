package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naked Singularity — Artifact {5} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep {3}
//	 If tapped for mana, Plains produce {R}, Islands produce {G}, Swamps produce
//	 {W}, Mountains produce {U}, and Forests produce {B} instead of any other
//	 type."
//
// Contamination's replacement with an answer per basic land type
// (OldB839LandTypeSwap). It changes the type of each mana and not the amount; a
// land with two of the types is answered per mana by the type its colour belongs
// to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb5ec55d-4a35-432a-be6e-c295f1b2e603",
		Name:         "Naked Singularity",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Naked Singularity — cumulative upkeep {3}", "{3}")},
		Replacements: []game.ReplacementEffect{
			effects.OldB839LandTypeSwap("Naked Singularity: Plains produce {R}, Islands {G}, Swamps {W}, Mountains {U} and Forests {B}"),
		},
	})
}
