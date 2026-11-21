package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glitterflitter — Creature — Faerie Performer {2}{U}, 2/2:
//
//	"Flying
//	 When Glitterflitter enters, you get {TK}, then you may put a sticker on a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on (CR
// 123.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3786047e-f136-49f8-81d6-08beb3f9bcfd",
		Name:            "Glitterflitter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.WhenThisEntersGetTicketsThenSticker("Glitterflitter", 1)},
	})
}
