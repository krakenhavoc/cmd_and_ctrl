package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_that_creature_combat_damage.go — "<for a while>, whenever that
// creature deals combat damage to a player, <do something>" (#4455):
//
//	"Whenever either of those creatures deals combat damage to a player
//	 this combat, put a +1/+1 counter on it."                Love on the Battlefield
//	"Until your next turn, whenever any of those creatures deals combat
//	 damage to a player, draw a card."                       Kang Dynasty
//	"Until end of turn, whenever he deals combat damage to an opponent,
//	 he deals that much damage to each other opponent."      Human Torch
//
// A repeating delayed triggered ability the resolving ability creates (CR
// 603.7a), one for each creature the clause names, so "either/any of those
// creatures" is one watch each. It is an each-condition like dies/that-creature
// (delayed_that_creature_dies.go), keyed by CondParams.Object, the creature's
// instance ID and CR 400.7 epoch: a creature that left and came back is a new
// object and never fires it (CR 603.7c). The epoch of the creature that dealt
// the damage is read from the last-known record, so a creature that dealt
// lethal-to-itself damage in the same step still triggers (CR 603.10a).
// Captain Howler's thatCreatureHitAPlayer is the older, epoch-less shape for a
// turn-long watch.
//
// The fired instance's payload is the creature with its epoch, and
// ctx.Trigger().Event is the damage event ("that much" is its Amount, the
// damaged player its Target).

var (
	// "Whenever that creature deals combat damage to a player".
	thatCreatureHitsAPlayerCondition = game.DelayedEachCondition("damage/that-creature-combat-to-player",
		thatCreatureCombatDamage(false))

	// "Whenever that creature deals combat damage to an opponent": a player
	// other than the delayed trigger's controller.
	thatCreatureHitsAnOpponentCondition = game.DelayedEachCondition("damage/that-creature-combat-to-opponent",
		thatCreatureCombatDamage(true))

	// "Put a +1/+1 counter on it": the payload creature, while it is still that
	// object on the battlefield.
	plusOneCounterOnThePayloadBody = game.SimpleDelayedBody("counter/plus-one-on-the-payload-creature", plusOneCounterOnThePayload)

	// "It deals that much damage to each other opponent": the payload creature
	// is the source, the damaged player is skipped.
	thatMuchToEachOtherOpponentBody = game.SimpleDelayedBody("damage/that-much-to-each-other-opponent", thatMuchToEachOtherOpponent)
)

func thatCreatureCombatDamage(opponentOnly bool) game.EachConditionFunc {
	return func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || p.Object.ID == uuid.Nil || ev.Source != p.Object.ID {
			return nil
		}
		if g.PlayerByIDForEffect(ev.Target) == nil || (opponentOnly && ev.Target == dt.Controller) {
			return nil
		}
		ref, ok := g.PermanentRefForEffect(ev.Source)
		if !ok || ref.Epoch != p.Object.Epoch {
			return nil
		}
		return []uuid.UUID{ev.Source}
	}
}

// whenThatCreatureHitsAPlayer creates "for <until>, whenever <creature> deals
// combat damage to a player (or an opponent), <body>" about the creature as
// the object it is now, controlled by the resolving object's controller (CR
// 603.7d). A creature no longer on the battlefield creates nothing: the
// trigger could never fire.
func whenThatCreatureHitsAPlayer(ctx *Context, creature uuid.UUID, label string, opponentOnly bool, until game.Duration, body game.BodyRef) error {
	c, ok := battlefieldCardForEffect(ctx.Game, creature)
	if !ok || creature == uuid.Nil {
		return nil
	}
	cond := thatCreatureHitsAPlayerCondition
	if opponentOnly {
		cond = thatCreatureHitsAnOpponentCondition
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        label,
		On:           []game.EventKind{game.EventDealDamage},
		Condition:    cond,
		CondParams:   game.EffectParams{Object: game.ObjectRef{ID: creature, Epoch: c.ObjectEpoch}},
		Body:         body,
		Repeats:      true,
		Duration:     &until,
	})
	return nil
}

func plusOneCounterOnThePayload(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	c, ok := battlefieldCardForEffect(g, obj.ID)
	if !ok || c.ObjectEpoch != obj.Epoch {
		return nil
	}
	return AddCounter{Target: obj.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}

func thatMuchToEachOtherOpponent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj, ok := payloadObject(ctx)
	if !ok {
		return nil
	}
	ev := ctx.Trigger().Event
	// One printed instruction: every opponent's damage is one instance
	// (CR 615.8).
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if opp == ev.Target {
				continue
			}
			if err := (DealDamage{SourceObject: &obj, Source: obj.ID, Target: opp, Amount: ev.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
