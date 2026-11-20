package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Armorer — Creature — Human Soldier {1}{W}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Other creatures you control get +0/+1."
//
// A layer 7c anthem over the other creatures its controller controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b1214e9-1581-4008-b3d4-8b7af6b092fd",
		Name:         "Veteran Armorer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 0, 1)},
	})
}
