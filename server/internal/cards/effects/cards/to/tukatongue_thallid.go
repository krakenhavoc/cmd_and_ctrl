package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tukatongue Thallid — Creature — Fungus {G}, 1/1 (Foundations Jumpstart):
//
//	"When this creature dies, create a 1/1 green Saproling creature token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7dbd008-e941-497e-8305-80c53b1dffe2",
		Name:         "Tukatongue Thallid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisDies("Tukatongue Thallid — create a 1/1 green Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})), game.Purpose{Tokens: 1}),
		},
	})
}
