package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minotaur de Force — Creature — Minotaur Performer {1}{R}, 2/1:
//
//	"Haste
//	 When this creature enters, you get {TK}, then you may put a sticker on a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on (CR
// 123.3c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67151297-3d25-4882-b562-56025ee81f48",
		Name:            "Minotaur de Force",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{WhenThisEntersGetTicketsThenSticker("Minotaur de Force", 1)},
	})
}
