package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// minus_counter_payment.go — #5746: "unless <player> puts a -1/-1 counter
// on a creature they control" as a pay-unless payment, and the Fallen
// Empires Chants built on it.
//
//	"Whenever a player puts a Swamp onto the battlefield, this enchantment
//	 deals 3 damage to that player unless the player puts a -1/-1 counter
//	 on a creature they control." (Thelon's Chant)
//
// CR 118.12a: "[do something] unless [a player does something else]" is
// "[that player] may [do something else]. If [they don't], [do
// something]", and CR 118.12 makes the action a cost paid as the ability
// resolves. So it is the ordinary pay-unless prompt with a non-mana
// payment (game.PayActionMinusOneCounter): the payer names one creature
// they control and it gets the counter, or they decline and the
// consequence runs. The counter is a cost's, so Doubling Season does not
// double it (CR 614.16). Only a creature that can have a counter put on it
// may pay (CR 118.3), so a payer with none is not asked.

// MinusOneCounterPayment is "put a -1/-1 counter on a <noun> you control"
// as a payment (#5746): `noun` is the printed word ("creature"), `of` the
// permanents that may pay. The engine leaves out one that can't have
// counters put on it.
func MinusOneCounterPayment(noun string, of game.PermanentQuery) UpkeepPayment {
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionMinusOneCounter, Count: 1, Of: of},
		Words:  "Put a -1/-1 counter on " + articleFor(noun) + " " + noun + " you control",
	}
}

// PayActionUnless is "<consequence> unless <Chooser> <makes a non-mana
// payment>" for any payment that is not mana (MinusOneCounterPayment,
// SacrificePayment, DiscardPayment, LifePayment, …), asked of a chooser
// who need not be the asking effect's controller. It is PayUnless over a
// PayAction: the registered Else runs on "no" and on a "yes" the chooser
// cannot make (CR 118.3), and Carry is the plain data it is handed after
// the asking controller, so a table waiting on the payer is a restore
// point. A mana Payment is asked as PayUnless asks it.
type PayActionUnless struct {
	Chooser  uuid.UUID
	Payment  UpkeepPayment
	Question string
	Else     game.PayUnlessElse
	Carry    []uuid.UUID
}

func (p PayActionUnless) Apply(ctx *Context) error {
	words, action := p.Payment.prompt()
	return ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
		Chooser: p.Chooser, Source: ctx.Source(), Cost: words, Question: p.Question, Action: action,
		Then: p.Else, Carry: append([]uuid.UUID{ctx.Controller()}, p.Carry...),
		Asked: game.NewAskedItem(ctx.Item),
	})
}

// chantDamage is the Chants' decline: the Chant deals 3 damage to the
// player who would not pay, carried as the branch's one ID. An on-disk
// key: never renamed, never reused.
var chantDamage = PayUnlessElse("pay-unless/chant-3-damage-to-that-player", func(ctx *Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return chantDealDamage(ctx, r.Carry[0])
})

// chantDamageAmount is the damage both Chants deal.
const chantDamageAmount = 3

func chantDealDamage(ctx *Context, victim uuid.UUID) error {
	if p := ctx.Game.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: victim, Amount: chantDamageAmount}.Apply(ctx)
}

// APlayerPutAPermanentWithSubtype is "whenever a player puts a <subtype>
// onto the battlefield": an EventETB whose permanent has the subtype as it
// enters, whoever put it there and under whoever's control. The player who
// put it is the event's PutBy (chantPutter).
func APlayerPutAPermanentWithSubtype(subtype string) When {
	return func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventETB {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.HasSubtype(subtype)
	}
}

// chantPutter is the player who put the triggering permanent onto the
// battlefield: the event's PutBy, or, on an entry that names nobody (a
// sandbox move), the player it entered under (CR 110.2a's default).
func chantPutter(g *game.Game, item *game.StackItem) uuid.UUID {
	if item == nil || item.Trigger == nil {
		return uuid.Nil
	}
	ev := item.Trigger.Event
	if ev.PutBy != uuid.Nil {
		return ev.PutBy
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); ok {
		return c.Controller
	}
	return uuid.Nil
}

// chantLandPunisher is a Chant's second ability: "Whenever a player puts a
// <subtype> onto the battlefield, this enchantment deals 3 damage to that
// player unless the player puts a -1/-1 counter on a creature they
// control." "That player" is the one who put the land there, even if it
// entered under someone else (the 2004-10-04 ruling). The trigger goes on
// the stack whether or not they control a creature; as it resolves, one
// who controls none that could take the counter has no payment to make
// (CR 118.3) and simply takes the damage, without being asked. The damage
// is the Chant's, and is dealt even if the Chant has left the battlefield
// in response (its last known information).
func chantLandPunisher(name, subtype string) game.TriggeredAbility {
	label := name + " — 3 damage to the player who put a " + subtype + " onto the battlefield unless they put a -1/-1 counter on a creature they control"
	return On(game.EventETB, APlayerPutAPermanentWithSubtype(subtype), label, func(g *game.Game, item *game.StackItem) error {
		victim := chantPutter(g, item)
		if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
			return nil
		}
		ctx := NewContext(g, item)
		pay := MinusOneCounterPayment("creature", QueryTypes("creature"))
		if !g.PayActionCanPayForEffect(victim, pay.Action) {
			return chantDealDamage(ctx, victim)
		}
		return PayActionUnless{
			Chooser:  victim,
			Payment:  pay,
			Question: name + " — put a -1/-1 counter on a creature you control, or take 3 damage?",
			Else:     chantDamage,
			Carry:    []uuid.UUID{victim},
		}.Apply(ctx)
	})
}
