package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran Gargoyle — Creature — Gargoyle {5}{W}, 3/3 (Dominaria
// Remastered, #5948):
//
//	"Flying, first strike
//	 Whenever this creature deals damage, you gain that much life."
//
// Any damage it deals, combat (both strikes) or otherwise, is its own
// trigger carrying that damage's amount.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b82d1104-8a54-4745-ba8b-eeddcb657338",
		Name:            "Kjeldoran Gargoyle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.GapB139WheneverThisDealsDamageYouGainThatMuchLife("Kjeldoran Gargoyle"),
		},
	})
}
