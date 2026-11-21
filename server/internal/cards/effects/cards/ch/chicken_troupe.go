package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chicken Troupe — Creature — Bird Performer {1}{G}, 2/2:
//
//	"Ward {2}
//	 When Chicken Troupe enters, you get {TK}, then you may put a sticker
//	 on a nonland permanent you own."
//
// Ward is the engine's triggered ability (Ward); the enters trigger is the
// line most sticker cards print (WhenThisEntersGetTicketsThenSticker).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68982532-c0c8-477b-9dd4-180c6d22731e",
		Name:         "Chicken Troupe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Chicken Troupe — ward {2}"),
			effects.WhenThisEntersGetTicketsThenSticker("Chicken Troupe", 1),
		},
	})
}
