package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_creature_events.go — repeating delayed triggers about ANY
// creature (#3997): "until end of turn, whenever a creature is dealt
// damage, …" and "… whenever a creature dies, …" (Shriveling Rot).
//
// Both are each-conditions (game.DelayedEachCondition, ADR 0187
// amendment 2026-10-16) naming the one creature the event is about, so
// the fired instance carries it as its payload, recorded with its CR
// 400.7 epoch (ADR 0026 amendment 2026-10-16). The keys are registered in
// delayed_bodies.go.
//
// The engine emits one EventDealDamage per source, so a "dealt damage"
// trigger that sets DelayedOnEvent.OncePerBatch triggers once for each
// creature in each damage instance: two blockers hitting one creature are
// one occurrence (CR 603.2c, 120.4b, 510.2).

// eachCreatureDealtDamage is the creature a damage event reached, while
// it is a creature on the battlefield. A damage event of 0 is no damage
// and never matches (CR 120.8).
func eachCreatureDealtDamage(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	c, ok := b28CreatureWasDealtDamage(ev, g)
	if !ok {
		return nil
	}
	return []uuid.UUID{c.InstanceID}
}

// eachCreatureDied is the creature an EventLTB to a graveyard names, if
// it was a creature as it last existed on the battlefield (CR 700.4,
// 603.10a).
func eachCreatureDied(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
	if _, ok := diedCreature(ev, g); !ok {
		return nil
	}
	return []uuid.UUID{ev.CardID}
}

// payloadObject is the fired instance's object: the first card on its
// payload, with the epoch the dispatch recorded.
func payloadObject(ctx *Context) (game.ObjectRef, bool) {
	for _, ref := range ctx.Payload() {
		if ref.Kind == game.TargetCard && ref.ID != uuid.Nil {
			return ref.Object(), true
		}
	}
	return game.ObjectRef{}, false
}

// destroyThePayloadPermanent is "destroy it": the payload's permanent, if
// it is still that object on the battlefield (CR 603.7c). One that died to
// the damage, or left and came back as a new object, is not affected.
func destroyThePayloadPermanent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, obj.ID)
	if !ok || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	return DestroyTarget{Target: obj.ID}.Apply(ctx)
}

// diedControllerLosesLifeEqualToToughness is "that creature's controller
// loses life equal to its toughness", both read off the payload's object
// as it last existed on the battlefield (CR 608.2h): the controller then,
// and the toughness with its counters and effects. A toughness of 0 or
// less loses no life.
func diedControllerLosesLifeEqualToToughness(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	info, ok := g.PermanentForEffect(obj)
	if !ok || info.Controller == uuid.Nil || info.Toughness <= 0 {
		return nil
	}
	return GainLife{Player: info.Controller, Amount: -info.Toughness}.Apply(ctx)
}
