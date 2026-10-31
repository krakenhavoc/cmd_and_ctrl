package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crimson Operative — Artifact Creature — Human Villain {3}{R}, 3/2
// (Marvel Super Heroes Commander, #4455):
//
//	"Prowess
//	 When this creature enters, exile the top card of your library.
//	 Until the end of your next turn, you may play that card."
//
// Prowess is the printed keyword (Agent of Atlas). The enters trigger is
// the shared impulse exile of one card through the end of the
// controller's next turn (Blazing Crescendo, Burning Curiosity).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8863742-6c1c-415a-ad27-7def6d119c6a",
		Name:            "Crimson Operative",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crimson Operative — exile the top card; you may play it until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
	})
}
