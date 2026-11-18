package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bladestitched Skaab — Creature — Zombie Soldier {U}{B}, 2/3 (Midnight
// Hunt):
//
//	"Other Zombies you control get +1/+0."
//
// Lord Of the Accursed's TribalAnthem, power only: other, yours, read
// through the layered view so a changeling or a granted Zombie type
// counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d72b8254-3df8-431b-955a-ec2aea493e2b",
		Name:         "Bladestitched Skaab",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 0),
		},
	})
}
