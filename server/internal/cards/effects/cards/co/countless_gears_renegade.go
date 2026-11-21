package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Countless Gears Renegade — Creature — Dwarf Artificer {1}{W}, 2/2:
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, create a 1/1 colorless
//	 Servo artifact creature token."
//
// Revolt is an intervening if (revolt.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d1f9036-a466-46b6-ab7c-2de487876978",
		Name:         "Countless Gears Renegade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersIfRevolt("Countless Gears Renegade — create a 1/1 colorless Servo artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
	})
}
