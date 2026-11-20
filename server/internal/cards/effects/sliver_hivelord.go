package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sliver Hivelord — Legendary Creature — Sliver {W}{U}{B}{R}{G}, 5/5
// (Magic 2015, #5556):
//
//	"Sliver creatures you control have indestructible. (Damage and
//	 effects that say "destroy" don't destroy them.)"
//
// Cloudshredder Sliver's layer-6 grant over "Sliver creatures you
// control" (YoursOnly), the Hivelord itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4fe33e7-b952-4152-8c83-81d948756d2f",
		Name:         "Sliver Hivelord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "indestructible"),
		},
	})
}
