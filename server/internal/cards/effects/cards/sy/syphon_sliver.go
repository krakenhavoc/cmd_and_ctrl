package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syphon Sliver — Creature — Sliver {2}{B}, 2/2 (Magic 2014, #5563):
//
//	"Sliver creatures you control have lifelink. (Damage dealt by a Sliver creature you control also causes you to gain that much life.)"
//
// Bonescythe Sliver's shape with lifelink: a layer 6 grant to every Sliver creature
// its controller controls, this one included, and to no opponent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61f18560-8f26-4d7b-a6f8-ec388766123b",
		Name:         "Syphon Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, "lifelink"),
		},
	})
}
