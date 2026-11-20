package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sergeant-at-Arms — Creature — Human Soldier {2}{W}, 2/3 (Dominaria, #5530):
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
//	 When this creature enters, if it was kicked, create two 1/1 white Soldier creature tokens."
//
// A kicked enters trigger (CR 603.4): two 1/1 white Soldiers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "2ea95e03-b95d-4ccf-afc6-1a26ac6cf354",
		Name:          "Sergeant-at-Arms",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Sergeant-at-Arms — kicked: create two 1/1 white Soldier creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 2})), game.Purpose{Tokens: 2}),
		},
	})
}
