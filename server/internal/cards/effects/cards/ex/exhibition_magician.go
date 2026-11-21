package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Exhibition Magician — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeSteps("Create a 1/1 green and white Citizen creature token.", nil,
			effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1}),
		effects.ModeSteps("Create a Treasure token.", nil, effects.CreateToken{Template: effects.TreasureToken(), N: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:     "758e68fd-76ba-467e-a756-c5daa7db16d5",
		Name:         "Exhibition Magician",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
