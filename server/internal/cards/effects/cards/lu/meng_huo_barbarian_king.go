package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meng Huo, Barbarian King — Legendary Creature — Human Barbarian Soldier {3}{G}{G}, 4/4
// (Portal Three Kingdoms, #5622):
//
//	"Other green creatures you control get +1/+1."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea7fdff3-9d45-4a4a-b7fe-84b334f04291",
		Name:         "Meng Huo, Barbarian King",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("G"), 1, 1),
		},
	})
}
