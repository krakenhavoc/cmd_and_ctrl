package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandsteppe Outcast — Creature — Human Warrior {2}{W}, 2/1 (Foundations
// Jumpstart, S82 BNS):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Create a 1/1 white Spirit creature token with flying."
//
// A modal enters trigger (CR 603.3c): the mode is chosen as the trigger goes
// on the stack, and neither bullet targets. The counter goes on this
// creature only while it is still the same object on the battlefield.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Sandsteppe Outcast — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeSteps("Put a +1/+1 counter on this creature.", nil, CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		ModeSteps("Create a 1/1 white Spirit creature token with flying.", nil,
			CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}),
	)
	Register(Spec{
		OracleID:     "fccd17b1-ca0f-4649-a002-5cbf9666b510",
		Name:         "Sandsteppe Outcast",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
