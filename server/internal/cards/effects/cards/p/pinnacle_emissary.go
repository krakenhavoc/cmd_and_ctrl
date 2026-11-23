package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pinnacle Emissary — Artifact Creature — Robot {1}{U}{R}, 3/3 (Edge of
// Eternities):
//
//	"Whenever you cast an artifact spell, create a 1/1 colorless Drone
//	 artifact creature token with flying and "This token can block only
//	 creatures with flying."
//	 Warp {U/R}"
//
// The cast trigger resolves before the spell (it is above it on the
// stack) and regardless of the spell's fate. The Drone is the printed
// token, its block rule a catalog template (DroneBlockerToken, the Faerie
// of Into the Fae Court's shape). Warp's cost is the hybrid {U/R}, payable
// with either colour (CR 702.185).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "522e8db5-857e-492f-afed-d60cd2ce3dcd",
		Name:             "Pinnacle Emissary",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{U/R}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Artifact(), "Pinnacle Emissary — create a 1/1 colorless Drone artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.DroneBlockerToken(), N: 1})),
		},
	})
}
