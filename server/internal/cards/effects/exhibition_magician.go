package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exhibition Magician — Creature — Human Wizard {2}{R}, 2/1 (Streets of
// New Capenna):
//
//	"When this creature enters, choose one —
//	 • Create a 1/1 green and white Citizen creature token.
//	 • Create a Treasure token."
//
// A modal enters trigger (CR 603.3c): the mode is chosen as the trigger goes
// on the stack; neither bullet targets.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Exhibition Magician — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeSteps("Create a 1/1 green and white Citizen creature token.", nil,
			CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}),
		ModeSteps("Create a Treasure token.", nil, CreateToken{Template: TreasureToken(), N: 1}),
	)
	Register(Spec{
		OracleID:     "758e68fd-76ba-467e-a756-c5daa7db16d5",
		Name:         "Exhibition Magician",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
