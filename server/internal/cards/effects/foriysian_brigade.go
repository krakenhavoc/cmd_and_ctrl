package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foriysian Brigade — Creature — Human Soldier {3}{W}, 2/4 (Weatherlight, #5629):
//
//	"This creature can block an additional creature each combat."
//
// Two-Headed Giant of Foriys's static: one more block for itself
// (multi_block.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7261dc6d-7669-4682-977b-1304f57a201d",
		Name:         "Foriysian Brigade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CanBlockAdditional(selfOnly, 1)},
	})
}
