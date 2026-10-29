package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ant-Man, Colony Commander — Legendary Creature — Human Rogue Hero
// {1}{G}{U}, 2/2 (Marvel Super Heroes):
//
//	"Whenever Ant-Man attacks, you may pay {1}. When you do, put a +1/+1
//	 counter on target creature.
//	 Whenever you put a +1/+1 counter on a creature, create a 1/1 green
//	 Insect creature token. This ability triggers only once each turn."
//
// The attack trigger is Terra, Herald of Hope's MayPay / CR 603.12 pair: the
// payment is the whole condition and the counter is a second triggered
// ability whose target is chosen when it goes on the stack, so the creature
// is picked after the mana is spent.
//
// The Insect trigger watches every +1/+1 placement by you on a creature,
// whoever controls the creature, including the one the first ability puts
// (a counter on any creature counts; removal and counters other players put
// do not). "Only once each turn" is the per-turn trigger record
// (b11TriggeredThisTurn, as Exemplar of Light reads it): a trigger that has
// been put on the stack this turn keeps every later placement from
// triggering, even if the first Insect never lands. Counters put on several
// creatures at once (proliferate) are several placements and still trigger
// it once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0ed9fbe-324d-467a-a326-c6f85a4c688a",
		Name:         "Ant-Man, Colony Commander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(antManPayLabel, antManPayOffer),
			On(game.EventCounterPlaced, antManCounterPlaced, antManInsectLabel,
				Do(CreateToken{Template: TokenCard("1/1 green Insect"), N: 1})),
		},
	})
}

const (
	antManPayLabel    = "Ant-Man, Colony Commander — pay {1} to put a +1/+1 counter on target creature"
	antManInsectLabel = "Ant-Man, Colony Commander — create a 1/1 green Insect"
)

// antManPayOffer asks for the {1}; the reflexive trigger lives in OnPay.
func antManPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Ant-Man, Colony Commander — pay {1} to put a +1/+1 counter on target creature?",
		Then:     antManPaid,
	}.Apply(NewContext(g, item))
}

// antManPaid is the paid half, keyed: the reflexive trigger is queued
// against the asking item.
var antManPaid = PayUnlessBranches("pay-unless/ant-man-colony-commander-counter",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Ant-Man, Colony Commander — put a +1/+1 counter on target creature",
			Body:  antManCounterBody,
		}.Apply(ctx)
	}, nil)

// antManCounterBody is "when you do, put a +1/+1 counter on target
// creature".
var antManCounterBody = game.ReflexiveBody("ant-man-colony-commander/counter", simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

// antManCounterPlaced is the When of the Insect trigger: the source's
// controller put +1/+1 counters on a creature, and this ability has not
// been put on the stack yet this turn.
func antManCounterPlaced(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Label != game.CounterPlusOne {
		return false
	}
	if _, ok := countersPutBy(ev, source.Controller, g, true); !ok {
		return false
	}
	if !antManIsCreature(g, ev.Target) {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, antManInsectLabel)
}

// antManIsCreature reports whether id is a creature on the battlefield.
func antManIsCreature(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsCreature()
}
