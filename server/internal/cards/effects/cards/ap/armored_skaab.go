package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Skaab — Creature — Zombie Warrior {2}{U}, 1/4 (Innistrad, #5572):
//
//	"When this creature enters, mill four cards."
//
// Carrion Grub's entry: an enters trigger on the stack that mills its
// controller four cards (CR 701.13).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74d85903-6404-41d9-bc00-81c6b2f61bb0",
		Name:         "Armored Skaab",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Armored Skaab — mill four cards", effects.Do(effects.MillCards{N: 4})),
		},
	})
}
