package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Covetous Dragon — Creature {4}{R}, 6/5:
//
//	"Flying
//	 When you control no artifacts, sacrifice this creature."
//
// ADR 0107 §1 (#1858). Flying is the printed keyword. The sacrifice is a
// CR 603.8 state trigger over the artifacts its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf7710cb-70cb-42fb-b840-cc4f385daa7a",
		Name:            "Covetous Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouControlNo(effects.QueryType("artifact"), "Covetous Dragon — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
