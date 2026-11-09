package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hecteyes — Creature — Ooze Horror {1}{B}, 1/1 (Final Fantasy, #4462):
//
//	"When this creature enters, each opponent discards a card."
//
// Arbiter of Woe's discard (eachOpponentDiscardsOne): each opponent
// chooses a card without revealing it and the cards are discarded at
// the same time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a11b4ed-430d-4ef8-b0c1-6a022f12448c",
		Name:         "Hecteyes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hecteyes — each opponent discards a card",
				Do(Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) }))),
		},
	})
}
