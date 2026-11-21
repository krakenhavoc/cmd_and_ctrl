package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const eagleLabel = "Eagle of Deliverance — put an indestructible counter on another target creature you control; draw a card if its power is 2 or less"

// Eagle of Deliverance — Creature — Bird Soldier {4}{W}{W}, 5/5 (The Lord of
// the Rings: Tales of Middle-earth Commander):
//
//	"Flying
//	 When this creature enters, put an indestructible counter on another
//	 target creature you control. Draw a card if that creature's power is 2
//	 or less."
//
// Flying is printed. The counter is a keyword counter (CR 122.1b); the
// power is read after it lands, as the ability resolves, through the
// layers. A target that left in response gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4446e59c-29f7-4722-92be-2625c1d3dab9",
		Name:            "Eagle of Deliverance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters(eagleLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterIndestructible, N: 1}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok || c.PowerForComparison() > 2 {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}), effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
