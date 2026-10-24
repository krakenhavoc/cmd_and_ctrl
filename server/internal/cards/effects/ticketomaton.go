package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ticketomaton — Artifact Creature — Robot {3}, 2/2:
//
//	"When Ticketomaton enters, you get {TK}, then you may put a sticker on
//	 a nonland permanent you own."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4): which permanent (or none), which sticker
// of the three sheets its controller plays with, and for a name sticker
// where in the name it goes. Its ticket cost is paid as it goes on (CR
// 123.3c), so the ticket this trigger gives can pay for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc8290d7-83b0-49ef-b0de-7ebb3633a149",
		Name:         "Ticketomaton",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{WhenThisEntersGetTicketsThenSticker("Ticketomaton", 1)},
	})
}
