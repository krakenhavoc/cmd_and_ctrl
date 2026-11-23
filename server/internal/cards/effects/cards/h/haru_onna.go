package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haru-Onna — Creature — Spirit {3}{G}, 2/1 (Saviors of Kamigawa, #5601):
//
//	"When this creature enters, draw a card.
//	 Whenever you cast a Spirit or Arcane spell, you may return this
//	 creature to its owner's hand."
//
// The enters trigger uses the stack. The return is the cycle's
// shared sokB1OnnaReturn: optional, asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fead5048-3ce5-46d9-9c04-0a52f1554030",
		Name:         "Haru-Onna",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Haru-Onna — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.SokB1OnnaReturn("Haru-Onna"),
		},
	})
}
