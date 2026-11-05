package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silicate Surveyor — Creature — Tholian Scout {4}{R}, 4/4 (Star Trek,
// #4973):
//
//	"Reach
//	 When this creature enters, exile the top card of your library.
//	 Until the end of your next turn, you may play that card."
//
// Reach is the printed keyword. The enters trigger is the shared impulse
// exile of one card through the end of the controller's next turn
// (Crimson Operative, Blazing Crescendo).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1b248ef5-ec57-481d-b47a-357d3867054f",
		Name:            "Silicate Surveyor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silicate Surveyor — exile the top card; you may play it until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
	})
}
