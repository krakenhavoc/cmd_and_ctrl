package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mental Note — Instant {U} (Judgment, #5611):
//
//	"Mill two cards.
//	 Draw a card."
//
// Two steps in printed order (CR 608.2c): the caster mills two (CR
// 701.17a; a short library mills what it has, CR 701.17b), then draws.
// Its home is a threshold deck, where the two cards are the point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8d5f31c-7abf-4fbb-977e-8353a97daf7a",
		Name:         "Mental Note",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1}, // a cantrip answers nothing on the stack
		OnResolve:    ResolveSteps(MillCards{N: 2}, DrawCards{N: 1}),
	})
}
