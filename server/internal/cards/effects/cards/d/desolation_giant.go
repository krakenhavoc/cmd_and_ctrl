package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desolation Giant — Creature — Giant {2}{R}{R}, 3/3 (Commander 2015, S82 OLD
// #5546):
//
//	"Kicker {W}{W} (You may pay an additional {W}{W} as you cast this spell.)
//	 When this creature enters, destroy all other creatures you control. If it
//	 was kicked, destroy all other creatures instead."
//
// Whether the Giant was kicked is read off the permanent as the trigger goes
// on the stack (CR 400.7d, game.CardKickedTimes) and stamped on the item, so
// a Giant that has left by resolution still destroys what it was kicked to
// destroy. Kicked, every other creature is destroyed, yours and theirs,
// instead of only yours; the Giant itself is never destroyed by this.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Desolation Giant — destroy all other creatures you control; if it was kicked, destroy all other creatures instead",
		desolationGiantDestroy)
	enters.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, "Desolation Giant — destroy all other creatures you control; if it was kicked, destroy all other creatures instead")
		item.Params.Amount = game.CardKickedTimes(*source)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:      "c23aaf6d-151e-481a-a345-ca00c14940d1",
		Name:          "Desolation Giant",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}{W}")},
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered:     []game.TriggeredAbility{enters},
	})
}

func desolationGiantDestroy(g *game.Game, item *game.StackItem) error {
	other := effects.OtherThan(item.SourceCardID)
	if item.Params.Amount > 0 {
		return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), other)}.Apply(effects.NewContext(g, item))
	}
	return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), other, effects.YouControl())}.Apply(effects.NewContext(g, item))
}
