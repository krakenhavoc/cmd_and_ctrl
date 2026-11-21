package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tethered Griffin — Creature {W}, 2/3:
//
//	"Flying
//	 When you control no enchantments, sacrifice this creature."
//
// ADR 0107 §1 (#1858). Flying is the printed keyword. The sacrifice is a
// CR 603.8 state trigger over the enchantments its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "909e1bff-237a-4259-9c0f-419185681782",
		Name:            "Tethered Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNo(effects.QueryType("enchantment"), "Tethered Griffin — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
