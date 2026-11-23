package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82Bro10ReturnTargetToBattlefield is "return target … card from your
// graveyard to the battlefield" for the first legal target: the card comes
// back under its owner's control, which is the caster's, since the clause
// names your graveyard. With plusOneIfCreature it is Recommission's "If a
// creature enters this way, it enters with an additional +1/+1 counter on
// it": the card's kind is read before it moves, and the counter is an
// "enters with" clause (CR 614.1c) that rides the entry itself.
func S82Bro10ReturnTargetToBattlefield(plusOneIfCreature bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			var counters map[string]int
			if c, ok := g.LookupCardForEffect(t.ID); ok && plusOneIfCreature && c.IsCreature() {
				counters = map[string]int{game.CounterPlusOne: 1}
			}
			_, err := g.ReturnFromGraveyardWithCountersForEffect(t.ID, ctx.Controller(), false, counters)
			if err != nil && !errors.Is(err, game.ErrCardNotFound) {
				return err
			}
			return nil
		}
		return nil
	}
}

// thisOrAnotherLegendaryCreatureYouControlEntered is "whenever ~ or another
// legendary creature you control enters" (Frodo Baggins, Loran): the source's
// own entry, or a legendary creature entering under its controller's control.
func ThisOrAnotherLegendaryCreatureYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature() && c.IsLegendary()
}
