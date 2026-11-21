package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Armorguard — Creature — Cat Soldier {2}{G}{W}, 3/3 (Alara
// Reborn, #5586):
//
//	"When this creature enters, creatures you control get +1/+1 until end
//	 of turn."
//
// Inspiring Captain's enters trigger: the creatures are the ones you
// control as it resolves (CR 611.2c), and the bonus ends in the cleanup
// step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9738f9b7-f281-44f6-a746-9c83380b35b7",
		Name:         "Leonin Armorguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Leonin Armorguard — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Power: 1, Toughness: 1,
						Label: "Leonin Armorguard — +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
