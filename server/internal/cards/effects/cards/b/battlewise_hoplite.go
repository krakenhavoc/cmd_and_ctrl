package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battlewise Hoplite — Creature — Human Soldier {W}{U}, 2/2 (Theros, S82 OLD
// b388):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put a +1/+1
//	 counter on this creature, then scry 1."
//
// Favored Hoplite's counter (a Hoplite that left and came back is a new object
// and gets none, CR 400.7) and then the scry, which waits for the counter's
// prompt if a replacement asks one (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "416de3dd-c15d-4596-9386-5cac53302724",
		Name:         "Battlewise Hoplite",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Battlewise Hoplite — put a +1/+1 counter on it, then scry 1 (heroic)", effects.Do(
				effects.Instruction(effects.PlusOneCounterOnSourceIfSameObject),
				effects.Scry{N: 1},
			)),
		},
	})
}
