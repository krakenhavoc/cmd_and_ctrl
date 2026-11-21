package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patient Instructor — Creature — Human Citizen {2}{W/U}, 2/2:
//
//	"Vigilance
//	 When this creature enters, recruit. (Draw a card, then discard a
//	 card. If you discarded a nonland card, create a 1/1 white Human
//	 Soldier creature token.)"
//
// No simplification. The Hobbit (#4475).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bddd7e99-ec74-4ca6-9137-155b85695a95",
		Name:            "Patient Instructor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Patient Instructor — recruit", effects.Do(effects.Recruit{})),
		},
	})
}
