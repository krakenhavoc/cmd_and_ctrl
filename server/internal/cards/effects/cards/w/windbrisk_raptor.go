package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windbrisk Raptor — Creature — Bird {5}{W}{W}, 5/7 (Shadowmoor, #5591):
//
//	"Flying
//	 Attacking creatures you control have lifelink."
//
// Flying is a printed keyword. The grant is Berserkers' Onslaught's
// AttackingCreaturesHave: only a creature you control that is attacking
// has lifelink, and it ends when the creature leaves combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fd729f6-27ea-4b07-81f0-885f2c0c00eb",
		Name:            "Windbrisk Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking creatures you control have lifelink.", nil, "lifelink"),
		},
	})
}
