package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Captain — Creature — Human Knight {3}{W}, 3/3 (Innistrad
// Remastered):
//
//	"When this creature enters, creatures you control get +1/+1 until end
//	 of turn."
//
// Includes this creature; the set is fixed as the ability resolves
// (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a294001-645b-4691-bb32-c240ba5b5320",
		Name:         "Inspiring Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inspiring Captain — creatures you control get +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Power: 1, Toughness: 1,
						Label: "Inspiring Captain — +1/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
