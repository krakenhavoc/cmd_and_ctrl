package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Percussionist — Artifact Creature — Monkey Toy {R}, 1/1
// (Duskmourn, #4466):
//
//	"Haste
//	 When this creature dies, exile the top card of your library. You may
//	 play it until the end of your next turn."
//
// A dies trigger (ThisDied, read from the departure, CR 603.10a) whose
// effect is Crimson Operative's: exile the top card of its controller's
// library with a permission to play it (a land too, on a main phase with an
// empty stack, paying all costs) until the end of that player's next turn,
// the seat-turn duration (ADR 0063).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b74c8ac-9c19-45b0-935f-8630b8da4543",
		Name:            "Clockwork Percussionist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Clockwork Percussionist — exile the top card of your library; you may play it until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
	})
}
