package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riveteers Decoy — Creature — Human Warrior {1}{G}, 3/1:
//
//	"This creature must be blocked if able.
//	 Blitz {3}{G}"
//
// Blitz (CR 702.152, #3772) on the shared MustBeBlocked static (#1597,
// Canopy Stalker's).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1c4918e6-3810-458e-a11b-949bbc43e7cc",
		Name:             "Riveteers Decoy",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Blitz("{3}{G}")},
		Static:           []game.StaticAbility{MustBeBlocked()},
	})
}
