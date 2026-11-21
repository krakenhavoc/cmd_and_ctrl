package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Messenger Drake — Creature — Drake {3}{U}{U}, 3/3 (Magic 2014, #5563):
//
//	"Flying
//	 When this creature dies, draw a card."
//
// Kingfisher's dies trigger on a 3/3 flier: the player who last controlled it
// draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c62dce6-dc1f-4d2d-95dd-295810ada6a6",
		Name:            "Messenger Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Messenger Drake — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
