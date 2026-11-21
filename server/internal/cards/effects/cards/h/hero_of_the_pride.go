package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of the Pride — Creature — Cat Soldier {1}{W}, 2/2 (Theros Beyond Death, #5514):
//
//	"Whenever you cast a spell that targets this creature, creatures
//	 you control get +1/+0 until end of turn."
//
// Heroic is Hero of the Games' trigger: a spell its controller casts
// that targets it pumps every creature they control.
//
// No simplification.
func init() {
	const label = "Hero of the Pride — creatures you control get +1/+0 until end of turn (heroic)"
	effects.Register(effects.Spec{
		OracleID:     "f3a2d80c-de6f-4b0e-aeb1-78639d55a6e6",
		Name:         "Hero of the Pride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
