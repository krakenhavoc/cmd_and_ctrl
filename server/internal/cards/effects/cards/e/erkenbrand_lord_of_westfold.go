package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const erkenbrandLabel = "Erkenbrand, Lord of Westfold — creatures you control get +1/+0 until end of turn"

// Erkenbrand, Lord of Westfold — Legendary Creature — Human Soldier {3}{R},
// 3/3 (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Erkenbrand or another Human you control enters, creatures you
//	 control get +1/+0 until end of turn."
//
// Akoum Battlesinger's group pump (mandatory here) on a "this or another
// Human" enters watch. The creatures that get it are those its controller
// controls as the trigger resolves (CR 611.2c), so a creature that enters
// afterwards does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cac919f-3bb8-49cb-b4e4-4dd0dcd7746c",
		Name:         "Erkenbrand, Lord of Westfold",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.GapB126ThisOrAnotherEntered("Human"), erkenbrandLabel,
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: "Erkenbrand, Lord of Westfold — +1/+0"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
