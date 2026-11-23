package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kargan Warleader — Creature — Human Warrior {1}{R}{W}, 3/3 (Zendikar
// Rising, #5508):
//
//	"Other Warriors you control get +1/+1."
//
// Suki, Kyoshi Captain's lord (TribalAnthem, others, yours only), read
// through the layered view so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63324e72-e580-456a-91be-766c9f07b7b3",
		Name:         "Kargan Warleader",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Warrior"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
