package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundshaker Sliver — Creature — Sliver {6}{G}, 5/5 (Magic 2014, #5563):
//
//	"Sliver creatures you control have trample. (A creature with trample can deal excess combat damage to the player or planeswalker it's attacking.)"
//
// Bonescythe Sliver's shape with trample: a layer 6 grant to every Sliver creature
// its controller controls, this one included, and to no opponent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "106e343f-c6f4-4155-ad01-63e482d4d38e",
		Name:         "Groundshaker Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "trample"),
		},
	})
}
