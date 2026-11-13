package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obsessive Search — Instant {U} (Torment, #5612):
//
//	"Draw a card.
//	 Madness {U} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// A one-mana cantrip whose real job is to be discarded: madness is the
// engine's (CR 702.35a, #657), grown from the one string, so a looter or
// a discard cost exiles it and offers the {U} cast at once. The draw is
// the controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "116ee88e-9e13-4f21-bc6a-efa994b5bf75",
		Name:         "Obsessive Search",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1}, // a cantrip answers nothing on the stack
		Madness:      "{U}",
		OnResolve:    ResolveSteps(DrawCards{N: 1}),
	})
}
