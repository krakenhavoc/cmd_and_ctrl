package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d72b8254-3df8-431b-955a-ec2aea493e2b",
		Name:         "Bladestitched Skaab",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 0),
		},
	})
}
