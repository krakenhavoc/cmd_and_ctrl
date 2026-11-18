package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unhallowed Phalanx — Creature — Zombie Soldier {4}{B}, 1/13 (Innistrad:
// Crimson Vow):
//
//	"This creature enters tapped."
//
// A self replacement effect, as on a tapped land (CR 614.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94c98a3b-9c3e-497f-b954-8f2248ca01ac",
		Name:         "Unhallowed Phalanx",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
	})
}
