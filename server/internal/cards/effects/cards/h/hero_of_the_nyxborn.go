package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of the Nyxborn — Enchantment Creature — Human Soldier {1}{R}{W}, 2/2 (Theros Beyond Death, #5514):
//
//	"When this creature enters, create a 1/1 white Human Soldier
//	 creature token.
//	 Whenever you cast a spell that targets this creature, creatures
//	 you control get +1/+0 until end of turn."
//
// Honored Knight-Captain's enters token, then Hero of the Games' heroic
// pump.
//
// No simplification.
func init() {
	const label = "Hero of the Nyxborn — creatures you control get +1/+0 until end of turn (heroic)"
	effects.Register(effects.Spec{
		OracleID:     "19e483e7-a0a5-4164-a5ea-dab51c07a3d7",
		Name:         "Hero of the Nyxborn",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hero of the Nyxborn — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
			effects.Heroic(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
