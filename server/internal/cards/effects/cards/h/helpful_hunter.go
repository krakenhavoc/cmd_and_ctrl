package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helpful Hunter — Creature — Cat {1}{W}, 1/1 (EDHREC rank 1958):
//
//	"When this creature enters, draw a card."
//
// The two-mana cantrip Cat. An ETB draw on the stack, so it can be
// responded to and re-fires on a blink, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0864adb-e9aa-40b6-91ff-a0646193e887",
		Name:         "Helpful Hunter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Helpful Hunter — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
