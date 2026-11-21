package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quick Sliver — Creature — Sliver {1}{G}, 1/1 (Commander Masters):
//
//	"Flash
//	 Any player may cast Sliver spells as though they had flash."
//
// Flash is printed data. The second line is Tidal Barracuda's declared cast
// timing for every player, narrowed to Sliver spells (CR 702.8): derived
// from the battlefield on every query, so it ends when Quick Sliver leaves
// and stacks with a second one. Only Sliver spells are opened; an opponent
// gets no flash for anything else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "530a5740-041c-42df-8fe0-fcea87258038",
		Name:            "Quick Sliver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{{
			Timing:  game.TimingFlash,
			Affects: game.TimingAffectsEachPlayer,
			Filter:  game.PermissionFilter{CreatureType: "Sliver"},
			Label:   "Any player may cast Sliver spells as though they had flash.",
		}},
	})
}
