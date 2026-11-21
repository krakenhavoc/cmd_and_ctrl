package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of the Winds — Creature — Human Soldier {3}{W}, 1/4 (Theros Beyond Death, #5514):
//
//	"Flying
//	 Whenever you cast a spell that targets this creature, creatures
//	 you control get +1/+0 until end of turn."
//
// Flying is the printed keyword; heroic is Hero of the Games' trigger.
//
// No simplification.
func init() {
	const label = "Hero of the Winds — creatures you control get +1/+0 until end of turn (heroic)"
	effects.Register(effects.Spec{
		OracleID:        "711893a5-2f4c-4957-90f5-2b1d6492b175",
		Name:            "Hero of the Winds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Heroic(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
