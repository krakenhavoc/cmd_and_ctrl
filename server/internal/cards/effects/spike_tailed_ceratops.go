package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spike-Tailed Ceratops — Creature — Dinosaur {4}{G}, 4/4 (Ixalan, #5532):
//
//	"This creature can block an additional creature each combat."
//
// Foriysian Brigade's static: it can block one more attacker each
// combat (CR 509.1a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "399ea8a4-55f5-4f14-ad7a-7af306959496",
		Name:         "Spike-Tailed Ceratops",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
	})
}
