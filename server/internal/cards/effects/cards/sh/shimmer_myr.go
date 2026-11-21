package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shimmer Myr — Artifact Creature — Myr {3}, 2/2:
//
//	"Flash
//	 You may cast artifact spells as though they had flash."
//
// Mai and Zuko's artifact-spell flash timing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e110ee80-37af-4989-a5fb-95d51dc7f102",
		Name:            "Shimmer Myr",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CardType: "artifact"},
				"You may cast artifact spells as though they had flash."),
		},
	})
}
