package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runewing — Creature — Bird {3}{U}, 2/2 (Return to Ravnica, #5566):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Kingfisher's dies trigger: it goes on the stack once the Bird is in the
// graveyard, and the player who last controlled it draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cc300649-ae50-47e7-ad0f-ffbdffddb989",
		Name:            "Runewing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Runewing — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
