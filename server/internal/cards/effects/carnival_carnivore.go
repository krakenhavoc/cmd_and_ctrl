package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnival Carnivore — Creature — Alien Crab Horror {5}{B}, 4/6:
//
//	"Deathtouch
//	 When Carnival Carnivore enters, you get {TK}, then you may put a sticker on a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on (CR
// 123.3c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba5bc14b-db2c-4561-a6a4-a02443791260",
		Name:            "Carnival Carnivore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered:       []game.TriggeredAbility{WhenThisEntersGetTicketsThenSticker("Carnival Carnivore", 1)},
	})
}
