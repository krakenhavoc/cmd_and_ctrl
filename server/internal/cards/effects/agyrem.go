package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agyrem — Plane — Ravnica (#4870):
//
//	"Whenever a white creature dies, return it to the battlefield under its
//	 owner's control at the beginning of the next end step.
//	 Whenever a nonwhite creature dies, return it to its owner's hand at the
//	 beginning of the next end step.
//	 Whenever chaos ensues, creatures can't attack you until a player
//	 planeswalks."
//
// The colour is the creature's as it last existed on the battlefield (the
// ruling; leftAsColor). Each trigger schedules a delayed trigger for the next
// end step that returns the card only if it is still the same object in a
// graveyard when it resolves (the rulings, CR 400.7); it fires even after the
// plane is gone. A token dies without leaving a card, so there is nothing for
// the delayed trigger to return. The chaos ability gives every other player a
// "can't attack you" restriction that ends when the plane stops being face up
// (game.GrantCantAttackPlayerUntilPlaneswalkForEffect); planeswalkers
// the planar controller controls can still be attacked (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38f84e55-049c-441e-b4e2-1e207ab5dbe5",
		Name:         "Agyrem",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftAsColor(ev, dead, source.TextColor("W"))
			}, "Agyrem — return it to the battlefield under its owner's control at the beginning of the next end step",
				b6AgyremSchedule(agyremReturnToBattlefieldBody, "Agyrem — return the white creature to the battlefield under its owner's control")),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !leftAsColor(ev, dead, source.TextColor("W"))
			}, "Agyrem — return it to its owner's hand at the beginning of the next end step",
				b6AgyremSchedule(agyremReturnToHandBody, "Agyrem — return the nonwhite creature to its owner's hand")),
			WheneverChaosEnsues("Agyrem — creatures can't attack you until a player planeswalks",
				func(g *game.Game, item *game.StackItem) error {
					for _, p := range apnapPlayers(g) {
						g.GrantCantAttackPlayerUntilPlaneswalkForEffect(p, item.Controller, game.CantAttackScope{PlayerOnly: true},
							"Agyrem — creatures can't attack you until a player planeswalks", item.SourceCardID)
					}
					return nil
				}),
		},
	})
}

// agyremReturnToBattlefieldBody and agyremReturnToHandBody are the delayed
// ability's bodies: the card named as the object it was, if it is still that
// object in a graveyard. On-disk keys: never renamed, never reused.
var (
	agyremReturnToBattlefieldBody = game.DelayedBody("planes-b6/agyrem-return-to-battlefield", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		return b6ReturnObjectFromGraveyard(g, item, p, game.ZoneBattlefield)
	})
	agyremReturnToHandBody = game.DelayedBody("planes-b6/agyrem-return-to-hand", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		return b6ReturnObjectFromGraveyard(g, item, p, game.ZoneHand)
	})
)

// b6AgyremSchedule schedules `body` at the next end step for the creature
// card that just died, when it is in a graveyard as the object it now is.
func b6AgyremSchedule(body game.BodyRef, label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		id := item.Trigger.Event.CardID
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label:  label,
			Cards:  []uuid.UUID{id},
			Body:   body,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}.Apply(NewContext(g, item))
	}
}
