package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Striking Sliver — Creature — Sliver {R}, 1/1 (Magic 2014, #5563):
//
//	"Sliver creatures you control have first strike. (They deal combat damage before creatures without first strike.)"
//
// Bonescythe Sliver's shape with first strike: a layer 6 grant to every Sliver creature
// its controller controls, this one included, and to no opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a15599d-8b00-4e0f-ac0c-d05f07c1e50d",
		Name:         "Striking Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "first strike"),
		},
	})
}
