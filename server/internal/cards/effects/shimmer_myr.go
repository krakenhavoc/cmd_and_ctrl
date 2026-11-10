package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shimmer Myr — Artifact Creature — Myr {3}, 2/2:
//
//	"Flash
//	 You may cast artifact spells as though they had flash."
//
// Mai and Zuko's artifact-spell flash timing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e110ee80-37af-4989-a5fb-95d51dc7f102",
		Name:            "Shimmer Myr",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CardType: "artifact"},
				"You may cast artifact spells as though they had flash."),
		},
	})
}
