package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blur Sliver — Creature — Sliver {2}{R}, 2/2 (Magic 2014, #5563):
//
//	"Sliver creatures you control have haste. (They can attack and {T} as soon as they come under your control.)"
//
// Bonescythe Sliver's shape with haste: a layer 6 grant to every Sliver creature
// its controller controls, this one included, and to no opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0f05979-daba-43eb-b356-24284e7ca99c",
		Name:         "Blur Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "haste"),
		},
	})
}
