package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// energy_payment.go — ADR 0129 §3: paying energy while a spell or ability
// resolves.
//
// CR 118.12: "[A player] may [do something]. If [that player] [does,
// doesn't, or can't], [effect]." … "The action [do something] is a cost,
// paid when the spell or ability resolves." CR 118.12a reads "[do
// something] unless [a player does something else]" the same way. Every
// energy payment printed inside an effect is one of those, and there are
// two shapes:
//
//   - A FIXED amount ("you may pay {E}{E}. If you do", "sacrifice it
//     unless you pay {E}", "an amount of {E} equal to its mana value"):
//     the pay-unless prompt with a PayActionEnergy payment
//     (QueueEnergyPaymentForEffect). The answer is yes or no; nothing is
//     named. "Pay" is a payment only when the payer has the energy
//     (CR 118.3), so a short payer's "yes" is the decline, as an
//     unfundable mana "yes" is.
//   - A CHOSEN amount ("you may pay any amount of {E}", "one or more
//     {E}"): PendingChoicePayAmount, a number between a floor and the
//     payer's energy (QueuePayEnergyAmountForEffect, owner decision 3).
//
// Both pay through payEnergyLocked (energy_cost.go), the one path that
// pays energy, so the log, the layer version and a later "paid or lost
// this turn" tally see every payment the same way. Neither is waived by
// the permissive posture or Cast anyway (owner decision 4).
//
// A mandatory payment ("pay {E}{E}. If you can't", Greenbelt Rampager)
// is no question at all: the payer pays when they can, and the "can't"
// branch runs when they can't (PayEnergyIfAbleForEffect).

// PayActionEnergy is a pay-unless payment of Count energy counters
// (CR 107.14). Nothing is named: the answer is yes or no.
const PayActionEnergy PayActionKind = "energy"

// EnergyPayment is a fixed-amount energy payment asked of Chooser while
// an ability resolves: "you may pay {E}{E}. If you do, …" (OnPay) or
// "sacrifice it unless you pay {E}" (OnDecline).
type EnergyPayment struct {
	// Chooser is the player asked to pay. Only they can pay: the
	// counters come off them (CR 107.14).
	Chooser uuid.UUID
	// Source attributes the prompt to a card for the client.
	Source uuid.UUID
	// N is the amount. Zero is a payment too ("an amount of {E} equal
	// to" a mana value of 0): it is asked, and always payable.
	N int
	// Question is the dialog header.
	Question string
	// OnPay runs once the energy has been paid. OnDecline runs on "no",
	// on a "yes" the payer cannot fund, and when the payer has left
	// (CR 800.4f: the cost is not paid).
	OnPay, OnDecline func(g *Game) error
	// InThisStep holds the table in the current step until the answer
	// (PendingChoice.OwedInStep): set it when what hangs off the answer
	// is about the step in progress — an attack trigger's +1/+1 counter
	// must land before combat damage. The energy helpers in
	// internal/cards/effects always set it.
	InThisStep bool
}

// EnergyPaymentWords is the payment's printed words for the prompt's
// PayCost and the bot's move label: "{E}{E}", or "0 {E}" for nothing.
func EnergyPaymentWords(n int) string {
	if n <= 0 {
		return "0 {E}"
	}
	return strings.Repeat("{E}", n)
}

// QueueEnergyPaymentForEffect asks p.Chooser whether to pay p.N energy.
// It is the pay-unless prompt (CR 118.12, 118.12a) with a PayActionEnergy
// payment, so the gate, the departure rule (dropDecline) and the client's
// dialog are the ones every pay-unless has. A payer who has left is not
// asked and OnDecline runs.
//
// Caller must hold g.mu.
func (g *Game) QueueEnergyPaymentForEffect(p EnergyPayment) error {
	n := p.N
	if n < 0 {
		n = 0
	}
	var owed TurnStep
	if p.InThisStep {
		owed = g.currentTurnStepLocked()
	}
	return g.queuePayUnlessLocked(p.Chooser, p.Source, EnergyPaymentWords(n), p.Question,
		p.OnDecline, owed, uuid.Nil, nil,
		payUnlessExtras{onPay: p.OnPay, action: &PayAction{Kind: PayActionEnergy, Count: n}})
}

// PayEnergyIfAbleForEffect is a MANDATORY payment, "pay {E}{E}. If you
// can't, …" (CR 118.12: the clause checks whether the player "started to
// pay a mandatory cost"). There is nothing to ask: a payer with the
// energy pays it and `then` runs with paid true; a payer without it pays
// nothing (CR 118.3) and `then` runs with paid false.
//
// Caller must hold g.mu.
func (g *Game) PayEnergyIfAbleForEffect(payer, source uuid.UUID, n int, then func(g *Game, paid bool) error) error {
	paid := false
	if p := g.playerByIDLocked(payer); p != nil && !p.Eliminated && EnergyShortfall(p, n) == nil {
		if err := g.payEnergyLocked(payer, n, source); err != nil {
			return err
		}
		paid = true
	}
	if then == nil {
		return nil
	}
	return then(g, paid)
}

// energyPaymentPayableLocked reports whether `chooser` can pay a
// PayActionEnergy payment now (CR 118.3).
func (g *Game) energyPaymentPayableLocked(chooser uuid.UUID, a *PayAction) bool {
	p := g.playerByIDLocked(chooser)
	return p != nil && !p.Eliminated && EnergyShortfall(p, a.Count) == nil
}

// PendingChoicePayAmount asks a player how much energy to pay: "you may
// pay any amount of {E}" (Harnessed Lightning), "you may pay one or more
// {E}" (Pia Nalaar). Owner decision 3 of ADR 0129: a number between
// PayAmount.Min and PayAmount.Max, answered with `amount` and drawn as a
// stepper in the action dock. It blocks the table: the spell that asked
// is paused mid-resolution and the rest of the card hangs off the answer.
const PendingChoicePayAmount PendingChoiceKind = "pay_amount"

// PayAmountPrompt is the plain data of a pay_amount prompt, read by the
// protocol view, the legal-move enumerator and the bot.
type PayAmountPrompt struct {
	// Min is the smallest payment: 0 for "any amount", 1 for "one or
	// more" (CR 107.1b: never negative). Max is the payer's energy when
	// the prompt was asked. Zero is always an answer as well: "any
	// amount" includes none, and "you may pay one or more" may be
	// declined. So the answers are 0 and Min..Max.
	Min, Max int
	// Goal is the smallest amount that reaches the card's own threshold
	// — the target's toughness for Harnessed Lightning — or 0 when the
	// card names none. The enumerator offers it beside Min and Max, and
	// the heuristic pays it.
	Goal int
	// Unit says what each counter paid buys: PayAmountDamage,
	// PayAmountCounters, PayAmountCards, … Read by the heuristic.
	Unit string
}

// The units a pay_amount prompt declares (PayAmountPrompt.Unit).
const (
	PayAmountDamage   = "damage"
	PayAmountCounters = "counters"
	PayAmountCards    = "cards"
	PayAmountPower    = "power"
	PayAmountTax      = "tax"
	PayAmountOther    = "other"
)

// PayEnergyAmount is what QueuePayEnergyAmountForEffect is handed.
type PayEnergyAmount struct {
	Chooser, Source uuid.UUID
	// Min is 0 for "any amount" and 1 for "one or more".
	Min int
	// Goal and Unit are PayAmountPrompt's. A negative Goal is "as much
	// as you can": the goal becomes the payer's energy, for a card whose
	// every counter buys more of something the controller wants
	// (Rampaging Aetherhood's +1/+1 counters).
	Goal     int
	Unit     string
	Question string
	// Then runs once the energy has been paid, with the amount: 0 when
	// nothing was paid. It runs exactly once, also when there was
	// nothing to ask and when the chooser left.
	Then func(g *Game, paid int) error
}

// QueuePayEnergyAmountForEffect asks p.Chooser how much energy to pay.
// The ceiling is their energy now. When there is nothing to choose — no
// energy, or less than the floor (CR 118.3) — nobody is asked and Then
// runs with 0. A chooser who has left is not asked either.
//
// The continuation rides the chooseValueResume frame a resolution-time
// colour pick uses, with the amount in its decimal form, so the prompt
// adds no closure route (ADR 0041's ratchet).
//
// Caller must hold g.mu.
func (g *Game) QueuePayEnergyAmountForEffect(p PayEnergyAmount) error {
	then := p.Then
	run := func(g *Game, n int) error {
		if then == nil {
			return nil
		}
		return then(g, n)
	}
	minimum := p.Min
	if minimum < 0 {
		minimum = 0
	}
	pl := g.playerByIDLocked(p.Chooser)
	if pl == nil || pl.Eliminated {
		return run(g, 0)
	}
	maximum := PlayerEnergy(pl)
	if maximum == 0 || maximum < minimum {
		return run(g, 0)
	}
	goal := p.Goal
	if goal < 0 {
		goal = maximum
	}
	if goal < minimum || goal > maximum {
		goal = 0
	}
	if p.Unit == "" {
		p.Unit = PayAmountOther
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:      PendingChoicePayAmount,
		Chooser:   p.Chooser,
		Count:     1,
		Source:    p.Source,
		Reason:    p.Question,
		PayAmount: &PayAmountPrompt{Min: minimum, Max: maximum, Goal: goal, Unit: p.Unit},
		chooseValueResume: &chooseValueFrame{then: func(g *Game, value string) error {
			// "" is the drop (the chooser left, CR 800.4f): nothing
			// was paid. An answer is always a validated number.
			n, err := strconv.Atoi(value)
			if err != nil {
				n = 0
			}
			return run(g, n)
		}},
	})
	if id == uuid.Nil {
		return run(g, 0)
	}
	return nil
}

// ResolvePayAmount answers a PendingChoicePayAmount with `amount`. The
// amount must be 0 or lie within the prompt's bounds, and within the
// chooser's energy now (CR 118.3); one that does not is refused with the prompt
// left in place. A legal answer pays the energy through payEnergyLocked
// and then runs the rest of the card with the amount.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolvePayAmount(choiceID, chooserID uuid.UUID, amount int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoicePayAmount || choice.PayAmount == nil {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	pa := choice.PayAmount
	if amount != 0 && (amount < pa.Min || amount > pa.Max) {
		return ErrInvalidParam
	}
	if err := EnergyShortfall(g.playerByIDLocked(chooserID), amount); err != nil {
		return err
	}
	frame := choice.chooseValueResume
	source := choice.Source
	g.dequeueChoiceLocked(idx)
	if err := g.payEnergyLocked(chooserID, amount, source); err != nil {
		g.emitChoiceEffectErrorLocked(chooserID, source, err)
		amount = 0
	}
	if frame != nil && frame.then != nil {
		if err := frame.then(g, strconv.Itoa(amount)); err != nil {
			g.emitChoiceEffectErrorLocked(chooserID, source, err)
		}
	}
	g.runStateChecksLocked()
	return nil
}
