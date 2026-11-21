package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusory Ambusher — Creature — Cat Illusion {4}{U}, 4/1 (Lost Caverns of
// Ixalan Commander):
//
//	"Flash
//	 Whenever this creature is dealt damage, draw that many cards."
//
// The trigger waits for the damage to finish and draws the TOTAL the
// Ambusher was dealt (CR 120.3), as Body of Knowledge does.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ad3add32-cb87-4cd4-8c36-a737a897a8c8",
		Name:            "Illusory Ambusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Illusory Ambusher — draw that many cards", effects.Cmdr2B08DrawThatMany),
		},
	})
}
