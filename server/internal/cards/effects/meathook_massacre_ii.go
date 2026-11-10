package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meathook Massacre II — Legendary Enchantment {X}{X}{B}{B}{B}{B}
// (Duskmourn, #4466):
//
//	"When Meathook Massacre II enters, each player sacrifices X creatures
//	 of their choice.
//	 Whenever a creature you control dies, you may pay 3 life. If you do,
//	 return that card under your control with a finality counter on it.
//	 Whenever a creature an opponent controls dies, they may pay 3 life.
//	 If they don't, return that card under your control with a finality
//	 counter on it."
//
// X is the X the spell was announced with (CR 107.3m, Card.CastX), read in
// Build and carried on Params.Amount as The Meathook Massacre does; a
// Massacre that entered without being cast has X = 0 and nothing is
// sacrificed. Every player, the controller included, names X creatures, one
// seat at a time in turn order, and they are sacrificed together (Ember
// Swallower's collector); a player with fewer creatures sacrifices what they
// have (CR 701.21a).
//
// "You" and "an opponent" are relative to the Massacre's controller, and
// "controls" is read off the controller the exit stamped on the event
// (leftUnderControlOf), so a stolen creature that dies pays out for its
// thief's side. Each dying creature is one trigger. The second ability is
// "you may pay 3 life" (MayPay, LifePayment: a controller who has fewer than
// 3 life is not asked, CR 119.4). The third makes the creature's controller
// choose (b23PayLifeUnless): they pay 3 life to keep it in the graveyard, and
// when they decline, or cannot pay, the card returns under the Massacre's
// controller's control. Either way the card comes back with a finality
// counter as it enters (CR 614.1c), whose exile replacement is the engine's
// (CR 122.1h, game/finality_counters.go). It returns only while it is still
// in a graveyard when the answer lands (the first and second rulings: a card
// that has left, and a token, which ceases to exist, do not return whatever
// was paid).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68957dca-df5c-4051-af22-407cb7e47200",
		Name:         "Meathook Massacre II",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Meathook Massacre II — each player sacrifices X creatures",
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Meathook Massacre II — each player sacrifices X creatures")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: meathookMassacreIIEdict,
			},
			TriggerWithPurpose(
				WheneverACreatureYouControlDies("Meathook Massacre II — you may pay 3 life; if you do, return that card under your control with a finality counter on it",
					meathookMassacreIIYou),
				game.Purpose{DeathPayoff: true}),
			TriggerWithPurpose(
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					return ok && leftUnderControlOf(ev, dead) != source.Controller
				}, "Meathook Massacre II — that creature's controller may pay 3 life; if they don't, return that card under your control with a finality counter on it",
					meathookMassacreIIOpponent),
				game.Purpose{DeathPayoff: true}),
		},
	})
}

// meathookMassacreIIEdict is the enters trigger's body: every seat in turn
// order from the active player is owed X creatures.
func meathookMassacreIIEdict(g *game.Game, item *game.StackItem) error {
	x := item.Params.Amount
	if x <= 0 {
		return nil
	}
	var asks []uuid.UUID
	if n := len(g.Seats); n > 0 {
		start := g.Turn.ActiveSeat
		for i := 0; i < n; i++ {
			p := g.Seats[(start+i)%n]
			if p == nil || p.Eliminated {
				continue
			}
			for k := 0; k < x; k++ {
				asks = append(asks, p.ID)
			}
		}
	}
	g.PlayersSacrificeForEffect(item.SourceCardID, asks, sacrificeSpec("a creature", Creature()), "Sacrifice a creature")
	return nil
}

// meathookMassacreIIYou is the second ability's body: the controller may pay
// 3 life, and the card returns once they have.
func meathookMassacreIIYou(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Pay:      LifePayment(3),
		Question: "Meathook Massacre II — pay 3 life to return that creature card under your control with a finality counter on it?",
		Then:     meathookMassacreIIPaid,
		Carry:    []uuid.UUID{item.Trigger.Event.CardID},
	}.Apply(NewContext(g, item))
}

// meathookMassacreIIPaid is "if you do, return that card under your control
// with a finality counter on it". An on-disk key: never renamed, never
// reused.
var meathookMassacreIIPaid = PayUnlessBranches("pay-unless/meathook-massacre-ii-you-paid",
	func(ctx *Context, r game.PayUnlessAnswer) error { return meathookMassacreIIReturn(ctx, r.Carry) }, nil)

// meathookMassacreIIOpponent is the third ability's body: the dead creature's
// controller is asked to pay 3 life.
func meathookMassacreIIOpponent(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	dead, ok := diedCreature(ev, g)
	if !ok {
		return nil
	}
	return b23PayLifeUnless{
		Chooser:  leftUnderControlOf(ev, dead),
		Life:     3,
		Question: "Meathook Massacre II — pay 3 life? If you don't, that creature card returns under Meathook Massacre II's controller's control with a finality counter on it.",
		Else:     meathookMassacreIIDeclined,
		Carry:    []uuid.UUID{ev.CardID},
	}.Apply(NewContext(g, item))
}

// meathookMassacreIIDeclined is "if they don't, return that card under your
// control with a finality counter on it". An on-disk key: never renamed,
// never reused.
var meathookMassacreIIDeclined = PayUnlessBranches("pay-unless/meathook-massacre-ii-opponent-declined",
	nil, func(ctx *Context, r game.PayUnlessAnswer) error { return meathookMassacreIIReturn(ctx, r.Carry) })

// meathookMassacreIIReturn returns the carried card from a graveyard to the
// battlefield under the asking controller's control with a finality counter,
// if it is still in a graveyard.
func meathookMassacreIIReturn(ctx *Context, carry []uuid.UUID) error {
	if len(carry) == 0 {
		return nil
	}
	id := carry[0]
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(id, ctx.Controller(), false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
}
