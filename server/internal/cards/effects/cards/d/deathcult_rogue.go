package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathcult Rogue — Creature — Human Rogue {1}{U/B}{U/B}, 2/2 (Gatecrash,
// #5565):
//
//	"This creature can't be blocked except by Rogues."
//
// Prowler's Helm's rule on itself with Rogues for Walls: "Rogues" is an
// effective creature type, so a changeling blocks it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa8ae8ac-4ad2-4cda-9f6b-4ec634526658",
		Name:         "Deathcult Rogue",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfCreatureType("Rogue"), "Rogues"),
		},
	})
}
