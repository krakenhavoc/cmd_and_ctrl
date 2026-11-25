package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wayward Guide-Beast — Creature — Beast {R}, 2/2 (Zendikar Rising, S82 OLD
// #5508):
//
//	"Trample, haste
//	 Whenever this creature deals combat damage to a player, return a land
//	 you control to its owner's hand."
//
// The land is the controller's choice as the trigger resolves, and mandatory
// (CR 609.3: nothing happens with no land to return).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c752e001-56cc-46dc-acb3-3fca6c0bdd2e",
		Name:            "Wayward Guide-Beast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Wayward Guide-Beast — return a land you control to its owner's hand",
				effects.Do(effects.ReturnOneYouControl{
					Match:    effects.MatchLand,
					Question: "Wayward Guide-Beast — return a land you control to its owner's hand",
				})),
		},
	})
}
