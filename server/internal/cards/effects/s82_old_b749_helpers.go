package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB749WheneverThisBlocksReturnItAtEndOfCombat is "Whenever this creature
// blocks a creature, return that creature to its owner's hand at end of
// combat." (Kaijin of the Vanishing Touch, Wall of Tears), named for the
// card in its labels.
//
// A block trigger once per attacker blocked (CR 509.3b), stamped with that
// attacker's object as it was blocked, that schedules a delayed trigger at the
// beginning of this combat's end of combat step (CR 511.2, 603.7) naming it.
// The delayed trigger is Wall of Junk's (ReturnObjectToHandBody): it returns
// the object only if it is still that object and still on the battlefield.
func OldB749WheneverThisBlocksReturnItAtEndOfCombat(card string) game.TriggeredAbility {
	label := card + " — return that creature to its owner's hand at end of combat"
	t := On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return B28SelfBlocked(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:     game.StepEndCombat,
			Label:  card + " — return the creature it blocked to its owner's hand",
			Body:   ReturnObjectToHandBody,
			Params: game.EffectParams{Object: item.Params.Object},
		}.Apply(NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	return t
}
