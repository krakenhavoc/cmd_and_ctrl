package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Percussionist — Artifact Creature — Monkey Toy {R}, 1/1
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 When this creature dies, exile the top card of your library. You
//	 may play it until the end of your next turn."
//
// Haste is the printed keyword. The dies trigger is Tasha Yar, Out of
// Time's (ExileTopNUntilYourNextTurn with one card): its controller is
// the one the Percussionist had as it died (CR 608.2h), and the play
// window runs through the end of that player's next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b74c8ac-9c19-45b0-935f-8630b8da4543",
		Name:            "Clockwork Percussionist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Clockwork Percussionist — exile the top card; you may play it until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 1) }))),
		},
	})
}
