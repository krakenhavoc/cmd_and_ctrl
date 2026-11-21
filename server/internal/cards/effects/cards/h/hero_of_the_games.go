package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of the Games — Creature — Human Soldier {2}{R}, 3/2 (Foundations
// Jumpstart, BNS batch 32):
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 creatures you control get +1/+0 until end of turn."
//
// The heroic trigger (shield_families_recipient.go) with Dawnhart Wardens'
// pump: the creatures it affects are fixed as the trigger resolves
// (CR 611.2c), Hero of the Games itself included.
//
// No simplification.
func init() {
	const label = "Hero of the Games — creatures you control get +1/+0 until end of turn (heroic)"
	effects.Register(effects.Spec{
		OracleID:     "fc594e38-4297-44d4-b11b-3ab804cfcbdd",
		Name:         "Hero of the Games",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
