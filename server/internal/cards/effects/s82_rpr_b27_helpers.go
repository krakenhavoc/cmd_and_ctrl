package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rprB27BloodthirstOnCast is "Whenever you cast a <spell> spell, it gains
// bloodthirst 3": a cast trigger that, on resolution, puts a bloodthirst
// mark (PromisedCounter.IfOpponentDealtDamage) on the spell it watched.
// The spell's identity is carried on the trigger as it is raised (the
// Wildgrowth Archaic shape), so a spell gone by resolution is not marked.
func rprB27BloodthirstOnCast(name string, spell CardPredicate, label string) game.TriggeredAbility {
	t := On(game.EventCast, YouCast(spell), label, func(g *game.Game, item *game.StackItem) error {
		g.MarkSpellEntersWithCountersForEffect(item.Params.Object, game.PromisedCounter{
			Kind:                  game.CounterPlusOne,
			Count:                 3,
			Source:                item.SourceCardID,
			SourceName:            name,
			IfOpponentDealtDamage: true,
		})
		return nil
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.CardID, Epoch: g.ObjectEpochForEffect(ev.CardID)}
		return item
	}
	return t
}
