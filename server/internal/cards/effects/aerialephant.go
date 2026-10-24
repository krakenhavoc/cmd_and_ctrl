package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerialephant — Creature — Elephant Performer {3}{W}, 3/2:
//
//	"Flying
//	 When Aerialephant enters, you get {TK}, then you may put a sticker on a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on (CR
// 123.3c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "20046568-b067-49fb-93b4-2ee86421f14b",
		Name:            "Aerialephant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{WhenThisEntersGetTicketsThenSticker("Aerialephant", 1)},
	})
}
