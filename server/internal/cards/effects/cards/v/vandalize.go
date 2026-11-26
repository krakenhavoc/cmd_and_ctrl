package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vandalize — Sorcery {4}{R} (Dragons of Tarkir):
//
//	"Choose one or both —
//	 • Destroy target artifact.
//	 • Destroy target land."
//
// Chosen both, each bullet has its own target and runs in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c71cbb3-e900-4ea3-aced-8439b5803654",
		Name:         "Vandalize",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target land.", effects.TargetPermanent("target land", effects.Land()), effects.DestroyTheModesTarget),
		),
	})
}
