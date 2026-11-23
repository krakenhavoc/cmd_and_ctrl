package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baloth Pup — Creature — Beast {1}{G}, 3/1 (Oath of the Gatewatch, #5545):
//
//	"This creature has trample as long as it has a +1/+1 counter on it."
//
// Chaos Imps' static: trample while the creature has a +1/+1 counter,
// judged continuously (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb06cf58-3cc6-429a-9bb0-be4b1c53de30",
		Name:         "Baloth Pup",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.ThisHasWhileItHasCounter(game.CounterPlusOne, "trample")},
	})
}
