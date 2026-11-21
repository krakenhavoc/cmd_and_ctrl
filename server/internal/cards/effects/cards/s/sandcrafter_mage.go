package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandcrafter Mage — Creature — Human Wizard {2}{W}, 2/2 (Dragons of Tarkir,
// #5551):
//
//	"When this creature enters, bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)"
//
// Bolster is effects.Bolster (CR 701.39): the creature with the least
// toughness is chosen as the trigger resolves, and does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bf9c91f-f278-4b97-8517-d062118c4e56",
		Name:         "Sandcrafter Mage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sandcrafter Mage — bolster 1", effects.Do(effects.Bolster{N: 1})),
		},
	})
}
