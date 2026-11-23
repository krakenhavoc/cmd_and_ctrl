package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sergeant-at-Arms — Creature — Human Soldier {2}{W}, 2/3 (Dominaria, #5530):
//
//	"Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
//	 When this creature enters, if it was kicked, create two 1/1 white Soldier creature tokens."
//
// A kicked enters trigger (CR 603.4): two 1/1 white Soldiers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "2ea95e03-b95d-4ccf-afc6-1a26ac6cf354",
		Name:          "Sergeant-at-Arms",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Sergeant-at-Arms — kicked: create two 1/1 white Soldier creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2})), game.Purpose{Tokens: 2}),
		},
	})
}
