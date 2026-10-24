package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stiltstrider — Creature — Rhino Performer {3}{G}{G}, 5/4:
//
//	"Reach
//	 When Stiltstrider enters, you get {TK}{TK}, then you may put a sticker on a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on (CR
// 123.3c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9c0d7d7d-5513-4e2b-b834-1ff999ab24f5",
		Name:            "Stiltstrider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{WhenThisEntersGetTicketsThenSticker("Stiltstrider", 2)},
	})
}
