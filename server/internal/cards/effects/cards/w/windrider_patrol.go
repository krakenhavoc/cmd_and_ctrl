package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Windrider Patrol — Creature — Merfolk Wizard {3}{U}{U}, 4/3 (Jumpstart 2022, #5948):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, scry 2."
//
// Flying rides PrintedKeywords; the trigger is the shared "this deals
// combat damage to a player" row with Octoprophet's scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dea0f359-19a5-4aef-8c96-574eaa748a31",
		Name:            "Windrider Patrol",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Windrider Patrol — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
