package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// plusCountersOnAnotherCreatureYouControl is the condition of "Whenever one
// or more +1/+1 counters are put on another creature you control" (Enduring
// Scalelord): the event PUT counters (a removal is not a placement) on a
// creature on the battlefield that its controller controls and that is not
// the source. The engine emits one EventCounterPlaced per permanent per
// placement, so "one or more" is the event itself.
func plusCountersOnAnotherCreatureYouControl(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Target == source.InstanceID || b33CountersPlacedDelta(ev, game.CounterPlusOne, g) <= 0 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !onBattlefield(g, ev.Target) {
		return false
	}
	return c.IsCreature() && c.Controller == source.Controller
}

// plusCountersOnThis is the condition of "Whenever one or more +1/+1
// counters are put on this creature" (Constable of the Realm): the event
// PUT counters on the source itself while it is on the battlefield.
func plusCountersOnThis(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.Target == source.InstanceID && onBattlefield(g, ev.Target) &&
		b33CountersPlacedDelta(ev, game.CounterPlusOne, g) > 0
}

// attackerWithPlusCounterYouControl is the condition of "Whenever a
// creature you control with a +1/+1 counter on it attacks" (Elite
// Scaleguard): the declared attacker is a creature its controller controls
// (the source's controller), and it has a +1/+1 counter on it as the
// attack is declared. The attacker may be the source itself.
func attackerWithPlusCounterYouControl(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && onBattlefield(g, ev.CardID) && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0
}

// targetCreatureTheAttackedPlayerControls is the target clause of "tap
// target creature defending player controls" for an attacker that is not the
// source: the defending player is the one the declared attacker (the
// event's CardID) attacks, read off the event's attack target.
func targetCreatureTheAttackedPlayerControls(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(tc.Event.Target)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "creature", Label: "target creature defending player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Controller == defender
		},
		Min: 1, Max: 1,
	}
}

// controlTwoOrMoreOfSubtype is the trigger-time half of the intervening "if
// you control two or more <land type>s" (Duergar Hedge-Mage), read for the
// source's controller through the layers as the event happens, for
// composing with AllOf. condControlTwoOrMoreOfSubtype is the resolution
// half (CR 603.4).
func controlTwoOrMoreOfSubtype(subtype string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return countControlled(g, source.Controller, func(c game.Card) bool { return c.HasSubtype(subtype) }) >= 2
	}
}

// condControlTwoOrMoreOfSubtype is the resolution recheck of the same
// condition: it counts the permanents the controller controls through the
// layers, so a land that is a Mountain by an effect counts.
func condControlTwoOrMoreOfSubtype(subtype string) CondBoard {
	return CondBoard{
		Name: "you control two or more " + subtype + "s",
		Test: func(ctx *Context) bool {
			return countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtype(subtype) }) >= 2
		},
	}
}
