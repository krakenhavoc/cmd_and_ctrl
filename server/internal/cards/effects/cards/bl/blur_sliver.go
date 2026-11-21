package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blur Sliver — Creature — Sliver {2}{R}, 2/2 (Magic 2014, #5563):
//
//	"Sliver creatures you control have haste. (They can attack and {T} as soon as they come under your control.)"
//
// Bonescythe Sliver's shape with haste: a layer 6 grant to every Sliver creature
// its controller controls, this one included, and to no opponent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0f05979-daba-43eb-b356-24284e7ca99c",
		Name:         "Blur Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "haste"),
		},
	})
}
