package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marchesa, the Black Rose — Legendary Creature — Human Wizard {1}{U}{B}{R},
// 3/3 (Conspiracy):
//
//	"Dethrone (Whenever this creature attacks the player with the most
//	 life or tied for most life, put a +1/+1 counter on it.)
//	 Other creatures you control have dethrone.
//	 Whenever a creature you control with a +1/+1 counter on it dies,
//	 return that card to the battlefield under your control at the
//	 beginning of the next end step."
//
// Dethrone is Scourge of the Throne's trigger; the grant gives it to every
// other creature you control (Nadu's grant shape), each with its own
// trigger. The death trigger reads the counters as the creature last
// existed (CR 603.10a), like Gladehart Cavalry's, and schedules a delayed
// trigger for the next end step (CR 603.7) naming the card as the object
// it is in the graveyard now (Nine-Lives Familiar's shape): a card that has
// left the graveyard by then is a new object and nothing returns (CR
// 400.7). It returns under Marchesa's controller's control. A token is not
// a card and returns nothing.
//
// No simplification.
func init() {
	const grantKey = "marchesa-the-black-rose/dethrone"
	const grantText = "Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)"
	effects.Register(effects.Spec{
		OracleID:     "17a59d3d-9e01-48cd-bb4a-3eaaa077751c",
		Name:         "Marchesa, the Black Rose",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{effects.GapB201Dethrone("dethrone: put a +1/+1 counter on it")},
			Text:      grantText,
		}},
		Static: []game.StaticAbility{effects.Labelled(effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
		}, grantKey), "Other creatures you control have dethrone.")},
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Marchesa, the Black Rose — dethrone: put a +1/+1 counter on it"),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Marchesa, the Black Rose — return that card to the battlefield at the beginning of the next end step", marchesaScheduleReturn),
		},
	})
}

// marchesaReturnBody is the delayed trigger's body: put the card back onto
// the battlefield under the controller's control if it is still the object
// that died, in a graveyard. An on-disk key: never renamed, never reused.
var marchesaReturnBody = game.DelayedBody("marchesa-the-black-rose/return", marchesaReturn)

func marchesaScheduleReturn(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Marchesa, the Black Rose — return that card to the battlefield",
		Cards:  []uuid.UUID{id},
		Body:   marchesaReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(effects.NewContext(g, item))
}

func marchesaReturn(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(id, item.Controller, false, nil)
	return effects.NothingIfGone(err)
}
