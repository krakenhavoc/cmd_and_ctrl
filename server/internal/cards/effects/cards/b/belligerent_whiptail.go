package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Belligerent Whiptail — Creature — Wurm {3}{R}, 4/2 (Battle for Zendikar):
//
//	"Landfall — Whenever a land you control enters, this creature gains first
//	 strike until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf0a15f5-5f41-4eea-81f7-02059b9586a2",
		Name:         "Belligerent Whiptail",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Belligerent Whiptail — gains first strike until end of turn (landfall)",
				effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Belligerent Whiptail — first strike")),
		},
	})
}
