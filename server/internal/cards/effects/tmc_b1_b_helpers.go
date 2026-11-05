package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b1_b_helpers.go — shared pieces for TMNT Commander batch 1,
// slice b (#5011). Append-only.

// tmcB1bThoseCreaturesDealtCombatDamage is "whenever one or more
// [creatures] you control deal combat damage to a player, put a +1/+1
// counter on each of those creatures and draw a card" for any `hit`
// condition: Vulture, Feathered Fiend's shape with the creature filter
// as a parameter.
//
// ONE trigger per damaged player and damage instance, settled once the
// combat damage step's events are all in (AtBatchEnd with
// ThatMuchDamage, CR 603.2c). The fill-in Build records "those
// creatures": the sources of this batch's combat damage to THAT player
// that `hit` accepted as the damage was dealt, each with its CR 400.7
// object epoch. A first-strike step and the regular step are two
// batches, so a creature that dealt damage only in the first step is
// not one of the second trigger's creatures. The resolution is
// vultureCountersAndDraw: a counter on each recorded creature still the
// same object on the battlefield, then one card.
func tmcB1bThoseCreaturesDealtCombatDamage(label string, hit When) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:    []game.EventKind{game.EventDealDamage},
		AppliesTo:  hit,
		Key:        label,
		AtBatchEnd: ThatMuchDamage(),
		Build: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			seen := map[game.ObjectRef]bool{}
			for i := len(g.Events) - 1; i >= 0; i-- {
				e := g.Events[i]
				if e.Batch < ev.Batch {
					break
				}
				if e.Batch != ev.Batch || e.Kind != game.EventDealDamage || e.Target != ev.Target || e.DamageInstance != ev.DamageInstance {
					continue
				}
				if !hit(e, source, lki, g) {
					continue
				}
				c, ok := g.LookupCardForEffect(e.Source)
				if !ok {
					continue
				}
				ref := game.ObjectRef{ID: e.Source, Epoch: c.ObjectEpoch}
				if !seen[ref] {
					seen[ref] = true
					item.Params.Objects = append(item.Params.Objects, ref)
				}
			}
			return item
		},
		Effect: vultureCountersAndDraw,
	}
}

// tmcB1bDoublesDrawing is Krang, the All-Powerful's "If a player drawing
// a card causes a triggered ability of a permanent you control to
// trigger, that ability triggers an additional time" (CR 603.2d).
//
// The cause is an EventDrawCard, which the engine emits once per card
// drawn (CR 121.2), by any player. The ability must be a permanent's
// (SourceIsPermanent): a spell's, an emblem's, a card's in a hand or a
// graveyard, and a delayed trigger (CR 603.2d) are not, and that
// permanent's controller must be Krang's.
func tmcB1bDoublesDrawing(name string) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    name,
		RowLabel: "Abilities of your permanents that trigger when a player draws a card trigger an additional time.",
		Applies: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.SourceIsPermanent || q.Ability == nil || q.Event.Kind != game.EventDrawCard {
				return false
			}
			return q.SourceLKI.Controller == q.DoublerLKI.Controller
		},
	}
}
